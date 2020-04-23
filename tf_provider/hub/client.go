package hub

import (
	"io"
	"io/ioutil"
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
}

// Service type contains the http client and its context info
type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment
}

// State contains the status of a membership
type State struct {
	Code string
}

//FIXME check the actual status code in the API and change if it does not match
const (
	StateREADY = "READY"
	StateNotPresent = "NOT_PRESENT"
)

func (s State) StateCode() string {
	return s.Code
}

// Endpoint contains a map with a membership's endpoint information
// At the moment it only has gke options
type Endpoint struct {
	GKECluster struct{
		ResourceLink string
	}
}

// Resource type contains specific info about a Hub membership resource
type Resource struct {
	// Name is the name of this membership. The name must be unique
	// within this project and zone, and can be up to 40 characters.
	Name string

	// Status is the current status of the membership. It could either be
	// StatusDone, StatusPending, StatusRunning, StatusError, StatusProvisioning, StatusStopping.
	State State

	// Endpoint is the url of the hub API.
	Endpoint Endpoint

	// Created is the creation time of this cluster.
	CreatedTime time.Time

	// Updated is the update time of this cluster.
	UpdatedTime time.Time
	
	// ExternalId is the uuid or the cluster name of the K8s cluster
	ExternalID string

	// ProjectID is the project id of this membership
	ProjectID string

	// Description is the description of this membership
	Description string
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
		c.Resource.State.Code = StateNotPresent
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
	c.Resource.ProjectID = c.projectID

	return nil
}

// CreateMembership updates a hub membership
// The client object should already contain the
// updated resource component updated in another method
func (c *Client) CreateMembership(ctx context.Context) error {
	// Try to populate the resource from the registry
	c.GetMembership(ctx, c.Resource.Name)
	if c.Resource.State.Code != StateNotPresent {
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
	u, err := url.Parse(APIURL)
	if err != nil {
		return retry.Unrecoverable(fmt.Errorf("Parsing %v url: %w", APIURL, err))
	}
	q := u.Query()
	q.Set("alt", "json")
	u.RawQuery = q.Encode()
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
