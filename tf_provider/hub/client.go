package hub

import (
	"io/ioutil"
	"time"
	"context"
	"fmt"
	"encoding/json"
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

// State contains the status of a membership
type State struct {
	Code string
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

	var resource Resource
	err = json.Unmarshal(body, &resource)
	if err != nil {
		return nil, fmt.Errorf("un-marshaling request body: %w", err)
	}
	resource.ProjectID = c.projectID

	return &resource, nil

}