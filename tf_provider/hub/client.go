package hub

import (
	"io"
	"io/ioutil"
	googletime "google.golang.org/genproto/googleapis/type/datetime"
	"time"
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
}

// Service type contains the http client and its context info
type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment
}

// MembershipState contains the status of a membership
type MembershipState struct {
	Code stateString `json:"code"`
	Description string `json:"description"` //Human readable description of the issue.\nThis field is deprecated, and is never set by the Hub Service.
	UpdateTime time.Time `json:"updateTime"`
}


type stateString string

// Code indicating the state of the Membership resource
const (
	MembershipStateCodeUnspecified stateString = "CODE_UNSPECIFIED"
	MembershipStateCreating = "CREATING" // CREATING indicates the cluster is being registered.
	MembershipStateReady = "READY" // READY indicates the cluster is registered.
	MembershipStateDeleting = "DELETING" // DELETING indicates that the cluster is being unregistered.
	MembershipStateUpdating = "UPDATING" // indicates the Membership is being updated.
	MembershipStateServiceUpdating = "SERVICE_UPDATING" // indicates the Membership is being updated by the Hub Service.
)

// MembershipEndpoint contains a map with a membership's endpoint information
// At the moment it only has gke options
type MembershipEndpoint struct {
	// If this Membership is a Kubernetes API server hosted on GKE, this is a
	// self link to its GCP resource.
	GKECluster GKECluster `json:"gkeCluster"`
}

// GKECluster represents a k8s cluster on GKE.
type GKECluster struct{
	// Self-link of the GCP resource for the GKE cluster. For example:
	// \/\/container.googleapis.com\/v1\/projects\/my-project\/zones\/us-west1-a\/clusters\/my-cluster
	// It can be at the most 1000 characters in length
	ResourceLink string `json:"resourceLink"` 
}

// Authority encodes how Google will recognize identities from this Membership.
// A workload with a token from this oidc_issuer can call the IAM credentials
// API for the provided identity_namespace and identity_provider; the workload
// will receive a Google OAuth token that it can use for further API calls.
// See the workload identity documentation for more details:
// https:\/\/cloud.google.com\/kubernetes-engine\/docs\/how-to\/workload-identity
type Authority struct {
	// An JWT issuer URI.\nGoogle will attempt OIDC discovery on this URI,
	// and allow valid OIDC tokens\nfrom this issuer to authenticate within
	// the below identity namespace.
	Issuer string `json:"Issuer"`

	// Output only. The identity namespace in which the issuer will be recognized.
	IdentityNamespace string `json:"identityNamespace"`

	// Output only. An identity provider that reflects this issuer in the identity namespace.
	IdentityProvider string `json:"identityProvider"`
}
var h Authority

// Resource type contains specific info about a Hub membership resource
type Resource struct {
	// Output only. The unique name of this domain resource in the format:
	// \n`projects\/[project_id]\/locations\/global\/memberships\/[membership_id]`.\n`membership_id`
	// can only be set at creation time using the `membership_id`\nfield in
	// the creation request. `membership_id` must be a valid RFC 1123\ncompliant
	// DNS label. In particular, it must be:\n  1. At most 63 characters in length\n  2. It must consist of lower case alphanumeric characters or `-`\n  3. It must start and end with an alphanumeric character\nI.e. `membership_id` must match the regex:
	// `[a-z0-9]([-a-z0-9]*[a-z0-9])?`\nwith at most 63 characters.
	Name string `json:"name"`

	// GCP labels for this membership."
	Labels string `json:"labels"`

	// Required. Description of this membership, limited to 63 characters.
	// It must match the regex: `a-zA-Z0-9*`
	Description string `json:"description"`

	Endpoint MembershipEndpoint `json:"endpoint"`

	// State is the current status of the membership
	State MembershipState `json:"state"`

	// How to identify workloads from this Membership.
	// See the documentation on workload identity for more details:
	// https:\/\/cloud.google.com\/kubernetes-engine\/docs\/how-to\/workload-identity
	Authority Authority `json:"authority"`

	// Output only. Timestamp for when the Membership was created.
	CreateTime googletime.DateTime `json:"createTime"`

	// Output only. Timestamp for when the Membership was last updated.
	UpdateTime googletime.DateTime `json:"updateTime"`

	//Output only. Timestamp for when the Membership was deleted.
	DeleteTime googletime.DateTime `json:"deleteTime"`
	
	// An externally-generated and managed ID for this Membership.
	// This ID may still be modified after creation but it is not
	// recommended to do so. The ID must match the regex: `a-zA-Z0-9*`
	ExternalID string `json:"externalId"`

	// Output only. For clusters using Connect, the timestamp
	// of the most recent connection established with Google Cloud.
	// This time is updated every several minutes, not continuously.
	// For clusters that do not use GKE Connect, or that have never
	// connected successfully, this field will be unset.
	LastConnectionTime string `json:"lastConnectionTime"`
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
func NewClient(ctx context.Context, projectID string) (*Client, error){
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

	//Populate the Client object itself
	c := &Client{
		projectID: projectID,
		svc: s,
		//FIXME not sure this should be hardcoded, but the api works as global, it will probably change in the future
		location: "global",
	}

	return c, nil
}

// GetMembership gets details of a hub membership.
// This method also initializes/updates the client component
func (c *Client) GetMembership(ctx context.Context, name string) error {
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

	if response.StatusCode == 404 {
		c.Resource.State.Code = MembershipStateCodeUnspecified
		return nil
	}

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad status code: %v", response.StatusCode)
	}

	err = json.Unmarshal(body, &c.Resource)
	if err != nil {
		return fmt.Errorf("un-marshaling request body: %w", err)
	}

	return nil
}

// CreateMembership updates a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) CreateMembership(ctx context.Context) error {
	err := c.ValidateExclusivity(ctx)
	if err != nil {
		return fmt.Errorf("Validating exclusivity: %w", err)
	}
	// Try to populate the resource from the registry
	c.GetMembership(ctx, c.Resource.Name)
	if c.Resource.State.Code != MembershipStateCodeUnspecified {
		return fmt.Errorf("Creating membership, the membership is already present")
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

// CallCreateMembershipAPI updates a hub membership
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

	fmt.Println(result)
	if result["done"] == true {
		return nil
	}
	return fmt.Errorf("not done")
}

// ValidateExclusivity checks the cluster exclusivity against the API
func (c *Client) ValidateExclusivity(ctx context.Context) error {
	// Call the gkehub api
	APIURL := prodAddr + "v1beta1/projects/" + c.projectID + "/locations/" + c.location + "/memberships:validateExclusivity"
	// Create the url parameters
	u, err := url.Parse(APIURL)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("Parsing %v url: %w", APIURL, err))
	}
	q := u.Query()
	q.Set("crManifest", c.K8S.CRManifest)
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
	//return fmt.Errorf("%v", u.String())
	// Go ahead with the request
	response, err := c.svc.client.Get(u.String())
	if err != nil {
		return fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading get request body: %w", err)
	}
	//DELETE ME
	return fmt.Errorf("%v", string(body))

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Bad status code: %v", string(body))
	}

	err = json.Unmarshal(body, &c.Resource)
	if err != nil {
		return fmt.Errorf("un-marshaling request body: %w", err)
	}

	return fmt.Errorf("Not an error: %v", string(body))
}