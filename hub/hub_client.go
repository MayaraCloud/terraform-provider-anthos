package hub

import (
	"context"
	"fmt"
	"net/http"

	"github.com/MayaraCloud/terraform-provider-anthos/k8s"
	"golang.org/x/oauth2/google"
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
	location  string // location of the membership
	Resource  Resource
	K8S       K8S
	ctx       context.Context
}

// K8S contains the membership K8S manifests
type K8S struct {
	CRManifest  string
	CRDManifest string
	Auth        k8s.Auth // K8s auth info
	UUID        string   // default namespace UID
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
func NewClient(ctx context.Context, projectID string, k8sAuth k8s.Auth) (*Client, error) {
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
		client:   httpClient,
		BasePath: endpoint,
	}

	// Populate the K8S object
	k := K8S{
		Auth: k8sAuth,
	}
	// Populate the Client object itself
	c := &Client{
		projectID: projectID,
		svc:       s,
		//FIXME not sure this should be hardcoded, but the api works as global, it will probably change in the future
		location: "global",
		K8S:      k,
		ctx:      ctx,
	}

	return c, nil
}

// GetKubeUUID grabs the namespace UID of the K8s cluster
func (c *Client) GetKubeUUID() error {
	kubeUUID, err := k8s.GetK8sClusterUUID(c.ctx, c.K8S.Auth)
	if err != nil {
		return fmt.Errorf("Getting uuid: %w", err)
	}
	c.K8S.UUID = kubeUUID
	return nil
}

// GetKubeArtifacts grabs the K8s CRD and manifest resource if existing
func (c *Client) GetKubeArtifacts() error {
	membershipCRD, err := k8s.GetMembershipCRD(c.ctx, c.K8S.Auth)
	if err != nil {
		return fmt.Errorf("Getting membership k8s crd: %w", err)
	}
	if membershipCRD != "" {
		membershipCR, err := k8s.GetMembershipCR(c.ctx, c.K8S.Auth)
		if err != nil {
			return fmt.Errorf("Getting membership k8s resource: %w", err)
		}
		c.K8S.CRManifest = membershipCR
	}
	c.K8S.CRDManifest = membershipCRD

	return nil
}
