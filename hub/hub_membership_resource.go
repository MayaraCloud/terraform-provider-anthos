package hub

import (
	"time"
)

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
	CreateTime time.Time `json:"createTime"`

	// Output only. Timestamp for when the Membership was last updated.
	UpdateTime time.Time `json:"updateTime"`

	//Output only. Timestamp for when the Membership was deleted.
	DeleteTime time.Time `json:"deleteTime"`

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

// MembershipState contains the status of a membership
type MembershipState struct {
	Code        stateString `json:"code"`
	Description string      `json:"description"` //Human readable description of the issue.\nThis field is deprecated, and is never set by the Hub Service.
	UpdateTime  time.Time   `json:"updateTime"`
}

type stateString string

// Code indicating the state of the Membership resource
const (
	MembershipStateCodeUnspecified stateString = "CODE_UNSPECIFIED"
	MembershipStateCreating                    = "CREATING"         // CREATING indicates the cluster is being registered.
	MembershipStateReady                       = "READY"            // READY indicates the cluster is registered.
	MembershipStateDeleting                    = "DELETING"         // DELETING indicates that the cluster is being unregistered.
	MembershipStateUpdating                    = "UPDATING"         // indicates the Membership is being updated.
	MembershipStateServiceUpdating             = "SERVICE_UPDATING" // indicates the Membership is being updated by the Hub Service.
)

// MembershipEndpoint contains a map with a membership's endpoint information
// At the moment it only has gke options
type MembershipEndpoint struct {
	// If this Membership is a Kubernetes API server hosted on GKE, this is a
	// self link to its GCP resource.
	GKECluster GKECluster `json:"gkeCluster"`
}

// GKECluster represents a k8s cluster on GKE.
type GKECluster struct {
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
