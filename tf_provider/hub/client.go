package hub

import (
	"io/ioutil"
	"time"
	"context"
	"fmt"
	"net/http"
	"google.golang.org/api/option"
	htransport "google.golang.org/api/transport/http"
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
}

// Status is the current status of the operation or resource.
type Status string

const (
	// StatusDone is a status indicating that the resource or operation is in done state.
	StatusDone = Status("done")
	// StatusPending is a status indicating that the resource or operation is in pending state.
	StatusPending = Status("pending")
	// StatusRunning is a status indicating that the resource or operation is in running state.
	StatusRunning = Status("running")
	// StatusError is a status indicating that the resource or operation is in error state.
	StatusError = Status("error")
	// StatusProvisioning is a status indicating that the resource or operation is in provisioning state.
	StatusProvisioning = Status("provisioning")
	// StatusStopping is a status indicating that the resource or operation is in stopping state.
	StatusStopping = Status("stopping")
)

// Resource type contains specific info about a Hub membership resource
type Resource struct {
	// Name is the name of this membership. The name must be unique
	// within this project and zone, and can be up to 40 characters.
	Name string

	// Description is the description of the membership. Optional.
	Description string

	// Status is the current status of the membership. It could either be
	// StatusDone, StatusPending, StatusRunning, StatusError, StatusProvisioning, StatusStopping.
	Status Status

	// Endpoint is the url of the hub API.
	Endpoint string

	// Created is the creation time of this cluster.
	Created time.Time

	// ProjectID is the project id of this membership
	ProjectID string
}

// NewClient creates a GKE hub client
func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error){
	// These are standard google api options
	o := []option.ClientOption{
		option.WithEndpoint(prodAddr),
		option.WithScopes(cloudPlatformScope),
		option.WithUserAgent(userAgent),
	}
	o = append(o, opts...)

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

// Service type contains the http client and its context info
type Service struct {
	client    *http.Client
	BasePath  string // API endpoint base URL
	UserAgent string // optional additional User-Agent fragment
}

// GetMembership gets details of a hub membership
func (c *Client) GetMembership(ctx context.Context, name string) (*Resource, error){
	// Call the gkehub api
	response, err := c.svc.client.Get(prodAddr + "v1/projects/" + c.projectID + "/locations/" + c.location + "/memberships/" + name)
	if err != nil {
		return nil, fmt.Errorf("get request: %w", err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading get request body: %w", err)
	}

	//FIXME this is still in development we need to properly parse the body json object in order to properly populate the Resource return object
	return &Resource{
		Name: string(body),
		Description: string(body),		
	}, nil

}