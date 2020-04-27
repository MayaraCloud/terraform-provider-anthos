package hub

import (
	"io"
	"io/ioutil"
	"context"
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	"bytes"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
	"github.com/avast/retry-go"
    "gitlab.com/mayara/private/anthos/k8s"
)

const prodAddr = "https://gkehub.googleapis.com/"
const userAgent = "gcloud-golang-hub/20200520"

const (
    // View and manage your data across Google Cloud Platform services
    cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"
)


// Client is a Google Connect Hub client, which may be used to manage
// hub memberships with a project. It must be constructed via NewClient.
type Client struct {
	projectID string
	svc       *Service
	location string // location of the membership
	Resource Resource
	K8S K8S
}

// K8S contains the membership K8S manifests
type K8S struct {
	CRManifest string
	CRDManifest string
	Auth k8s.Auth // K8s auth info
	UUID string // default namespace UID
}

// Service type contains the http client and its context info
type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment
}

// GetOptionsWithCreds initializes a GKEhub client object
func GetOptionsWithCreds(project string) (option.ClientOption, error) {
	// Get default credentials https://godoc.org/golang.org/x/oauth2/google
	creds, err := google.FindDefaultCredentials(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("Getting credentials: %w", err)
	}
	// Create google api options with the generated credentials
	options := option.WithCredentials(creds)

	return options, nil
}


// NewClient creates a GKE hub client
func NewClient(ctx context.Context, projectID string, k8sAuth k8s.Auth) (*Client, error){
	options, err := GetOptionsWithCreds(projectID)
	if err != nil {
		return nil, fmt.Errorf("Getting options with credentials: %w", err)
	}
	// These are standard google api options
	o := []option.ClientOption{
		option.WithEndpoint(prodAddr),
		option.WithScopes(cloudPlatformScope),
		option.WithUserAgent(userAgent),
	}
	o = append(o, options)

	// Create the client that actually makes the api REST requests
	httpClient, endpoint, err := htransport.NewClient(ctx, o...)
	if err != nil {
		return nil, fmt.Errorf("dialing: %v", err)
	}

	// Populate the svc object of the client
	s := &Service{
		client: httpClient,
		BasePath: endpoint,
	}

	// Populate the K8S object
	k := K8S{
		Auth: k8sAuth,
	}
	// Populate the Client object itself
	c := &Client{
		projectID: projectID,
		svc: s,
		//FIXME not sure this should be hardcoded, but the api works as global, it will probably change in the future
		location: "global",
		K8S: k,
	}

	return c, nil
}

// GetKubeUUID grabs the namespace UID of the K8s cluster 
func (c *Client) GetKubeUUID() error {
    kubeUUID, err := k8s.GetK8sClusterUUID(c.K8S.Auth)
    if err != nil {
        return fmt.Errorf("Getting uuid: %w", err)
	}
	c.K8S.UUID = kubeUUID
	return nil
}

// GetKubeArtifacts grabs the K8s CRD and manifest resource if existing
func (c *Client) GetKubeArtifacts() error {
    membershipCRD, err := k8s.GetMembershipCR(c.K8S.Auth)
    if err != nil {
        return fmt.Errorf("Getting membership k8s crd: %w", err)
	}
	if membershipCRD != "" {
		membershipCR, err := k8s.GetMembershipCR(c.K8S.Auth)
		if err != nil {
		    return fmt.Errorf("Getting membership k8s resource: %w", err)
		}
		c.K8S.CRManifest = membershipCR
	}
	c.K8S.CRDManifest = membershipCRD

	return nil
}


// GetMembership gets details of a hub membership.
// This method also initializes/updates the client component
func (c *Client) GetMembership(ctx context.Context, name string, checkNotExisting bool) error {
	// Call the gkehub api
	APIURL := prodAddr + "v1/projects/" + c.projectID + "/locations/" + c.location + "/memberships/" + name
	response, err := c.svc.client.Get(APIURL)
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}

	// If we are checking if the resource does not exist
	// we need a 404 here
	if checkNotExisting && response.StatusCode != 404 {
		return fmt.Errorf("The resource already exists in the Hub: %v", string(body))
	}

	if checkNotExisting && response.StatusCode == 404 {
		return nil
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad %v status code: %v", response.StatusCode, string(body))
	}

	err = json.Unmarshal(body, &c.Resource)
	if err != nil {
		return fmt.Errorf("un-marshaling request body: %w", err)
	}

	return nil
}

// CreateMembership creates a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) CreateMembership(ctx context.Context) error {
	// Validate exclusivity if the cluster has a manifest CRD present
	if c.K8S.CRManifest != "" {
		err := c.ValidateExclusivity(ctx)
		if err != nil {
			return fmt.Errorf("Validating exclusivity: %w", err)
		}
	}
	// Calling the creation API
	createResponse, err := c.CallCreateMembershipAPI(ctx)
	if err != nil {
		return fmt.Errorf("Calling CallCreateMembershipAPI: %w", err)
	}
	
	// Wait until we get an ok from CheckOperation
	retry.Attempts(60)
	err = retry.Do(
		func() error {
			return c.CheckOperation(ctx, createResponse["name"].(string))
		})

	if err != nil {
		return fmt.Errorf("Retry checking CreateMembership operation: %w", err)
	}
	return nil
}

// CallCreateMembershipAPI creates a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) CallCreateMembershipAPI(ctx context.Context) (HTTPResult, error) {
	// Create the json POST request body
	var rawBody struct{
		Description string `json:"description"`
		ExternalID string	`json:"externalId"`
	}
	rawBody.Description = c.Resource.Description
	rawBody.ExternalID = c.Resource.ExternalID

	body , err := json.Marshal(rawBody)
	if err != nil {
		return nil, fmt.Errorf("Marshaling create request body: %w", err)
	}
	// Create a url object to append parameters to it
	APIURL := prodAddr + "v1/projects/" + c.projectID + "/locations/" + c.location + "/memberships"
	u, err := url.Parse(APIURL)
	if err != nil {
		return nil, fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("alt", "json")
	q.Set("membershipId", c.Resource.Name)
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Post(u.String(), "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("create POST request: %w", err)
	}
	defer response.Body.Close()

	return DecodeHTTPResult(response.Body)
}

// HTTPResult is used to store the result of an http request
type HTTPResult map[string]interface{}

// DecodeHTTPResult decodes an http response body
func DecodeHTTPResult(httpBody io.ReadCloser) (HTTPResult, error) {
	var h HTTPResult
	err := json.NewDecoder(httpBody).Decode(&h)
	if err != nil {
		return nil, fmt.Errorf("Decoding http body response: %w", err)
	}
	return h, nil
}

// CheckOperation checks a hub operation status and returns true if the operation is done
func (c *Client) CheckOperation(ctx context.Context, operationName string) error {
	// Create a url object to append parameters to it
	APIURL := prodAddr + "v1/" + operationName
	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("Parsing %v url: %w", APIURL, err))
	}
	q := u.Query()
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("create POST request: %w", err))
	}
	defer response.Body.Close()
	
	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return retry.Unrecoverable(fmt.Errorf("Bad status code: %v", response.StatusCode))
	}
	
	result, err := DecodeHTTPResult(response.Body)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("Calling DecodeHTTPResult: %w", err))
	}

	if result["done"] == true {
		return nil
	}

	return fmt.Errorf("Failed to check operation: %v", result)
}

// ValidateExclusivity checks the cluster exclusivity against the API
func (c *Client) ValidateExclusivity(ctx context.Context) error {
	// Call the gkehub api
	APIURL := prodAddr + "v1beta1/projects/" + c.projectID + "/locations/" + c.location + "/memberships:validateExclusivity"
	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("crManifest", c.K8S.CRManifest)
	q.Set("intendedMembership", c.Resource.Name)
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad status code: %v", response.Body)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}
	var result GRCPResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("json Un-marshaling body: %w", err)
	}

	// 0 == OK in gRCP codes, see below.
	if result.Status.Code != 0 {
		return fmt.Errorf("%v", result.Status.Message)
	}

	return nil
}

// GRCPResponse follows the https://cloud.google.com/apis/design/errors
// Code must be one of the following
// https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
type GRCPResponse struct {
	Status GRCPResponseStatus `json:"status"`
}

// GRCPResponseStatus is the inner GRCPResponse struct
type GRCPResponseStatus struct {
	// Code contains the validation result. As such,
	// * OK means that exclusivity may be obtained if the manifest produced by
	// GenerateExclusivityManifest can successfully be applied.
	// * ALREADY_EXISTS means that the Membership CRD is already owned by another
	// Hub. See status.message for more information when this occurs
	Code int32 `json:"code"`
	Message string `json:"message"`
	Details map[string]interface{} `json:"details"`
}

// GenerateExclusivity checks the cluster exclusivity against the API
func (c *Client) GenerateExclusivity(ctx context.Context) error {
	// Call the gkehub api
	APIURL := prodAddr + "v1beta1/projects/" + c.projectID + "/locations/" + c.location + "/memberships/" + c.Resource.Name + ":generateExclusivity"

	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("crManifest", c.K8S.CRManifest)
	q.Set("crdManifest", c.K8S.CRDManifest)
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad status code: %v", response.Body)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}
	var result GRCPResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return fmt.Errorf("json Un-marshaling body: %w", err)
	}

	// 0 == OK in gRCP codes, see below.
	if result.Status.Code != 0 {
		return fmt.Errorf("%v", result.Status.Message)
	}

	return nil
}

// DeleteMembership deletes a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) DeleteMembership(ctx context.Context) error {
	// Calling the deletion API
	deleteResponse, err := c.CallDeleteMembershipAPI(ctx)
	if err != nil {
		return fmt.Errorf("Calling CallDeleteMembershipAPI: %w", err)
	}
	
	// Wait until we get an ok from CheckOperation
	retry.Attempts(60)
	err = retry.Do(
		func() error {
			return c.CheckOperation(ctx, deleteResponse["name"].(string))
		})

	if err != nil {
		return fmt.Errorf("Retry checking DeleteMembership operation: %w", err)
	}
	return nil
}

// CallDeleteMembershipAPI deletes a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) CallDeleteMembershipAPI(ctx context.Context) (HTTPResult, error) {
	// Delete a url object to append parameters to it
	APIURL := prodAddr + "v1/projects/" + c.projectID + "/locations/" + c.location + "/memberships/" + c.Resource.Name
	u, err := url.Parse(APIURL)
	if err != nil {
		return nil, fmt.Errorf("Parsing %v url: %w", APIURL, err)
	}
	q := u.Query()
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	// Go ahead with the request
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("Creating Delete request: %w", err)
	}
	response, err := c.svc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Sending DELETE request: %w", err)
	}
	defer response.Body.Close()

	return DecodeHTTPResult(response.Body)
}
