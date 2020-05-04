package hub

import (
	"context"
	"fmt"

	"github.com/MayaraCloud/terraform-provider-anthos/debug"
	"github.com/MayaraCloud/terraform-provider-anthos/k8s"
	"github.com/avast/retry-go"
)

// ParentRef is the resource name of the parent collection of a membership.
type ParentRef string

// GetParentRef gets the resource name of the parent collection of a membership.
func GetParentRef(project string, location string) ParentRef {
	return ParentRef(fmt.Sprintf("projects/%v/locations/%v", project, location))
}

// Global variable used for various context purposes
var ctx = context.Background()

// GetMembership gets a Membership resource from the GKEHub API
func GetMembership(project string, membershipID string, description string, gkeClusterSelfLink string, issuerURL string, k8sAuth k8s.Auth) error {
	client, err := NewClient(ctx, project, k8sAuth)
	if err != nil {
		return fmt.Errorf("Getting new client: %w", err)
	}
	return client.GetMembership(membershipID, false)
}

// CreateMembership creates a membership GKEHub resource
func CreateMembership(project string, membershipID string, description string, gkeClusterSelfLink string, issuerURL string, k8sAuth k8s.Auth) (membershipUUID string, err error) {
	client, err := NewClient(ctx, project, k8sAuth)
	if err != nil {
		return "", fmt.Errorf("Getting new client: %w", err)
	}

	// Get the K8s default namespace UID
	err = client.GetKubeUUID()
	if err != nil {
		return "", fmt.Errorf("Getting Kube UID: %w", err)
	}

	err = client.GetKubeArtifacts()
	if err != nil {
		return "", fmt.Errorf("Getting Kube custom artifacts: %w", err)
	}

	// Check if membership does not already exist
	err = client.GetMembership(membershipID, true)
	if err != nil {
		return "", fmt.Errorf("Checking if membership does not exist: %w", err)
	}

	// Populate the membership resource fields with the parameters
	client.Resource.Description = membershipID
	client.Resource.Endpoint.GKECluster.ResourceLink = gkeClusterSelfLink
	// Create the membership
	err = client.CreateMembership(membershipID)
	if err != nil {
		return "", fmt.Errorf("Creating membership membership: %w", err)
	}

	// Get membership info after creation, just to double check that all went fine
	err = client.GetMembership(membershipID, false)
	if err != nil {
		return "", fmt.Errorf("Checking getting membership info after creation: %w", err)
	}

	// Get Kubernetes artifacts to install or update the K8s CRD and CR
	err = client.GenerateExclusivity(membershipID)
	if err != nil {
		return "", fmt.Errorf("Generating K8s exclusivity artifacts: %w", err)
	}

	// Install the membership CRD and the membership CR in the kubernetes cluster
	err = k8s.InstallExclusivityManifests(client.ctx, k8sAuth, client.K8S.CRDManifest, client.K8S.CRManifest)
	if err != nil {
		return "", fmt.Errorf("Installing CRD and CR manifest in the Kubernetes cluster: %w", err)
	}

	return client.K8S.UUID, nil
}

// DeleteMembership deletes a membership GKEHub resource
func DeleteMembership(project string, membershipID string, description string, gkeClusterSelfLink string, issuerURL string, k8sAuth k8s.Auth, deleteArtifacts bool) error {
	client, err := NewClient(ctx, project, k8sAuth)
	if err != nil {
		return fmt.Errorf("Getting new client: %w", err)
	}
	// Get membership info
	err = client.GetMembership(membershipID, false)
	if err != nil {
		return fmt.Errorf("Checking membership info: %w", err)
	}

	// Delete the membership
	err = client.DeleteMembership()
	if err != nil {
		return fmt.Errorf("Deleting membership: %w", err)
	}

	// Wait until the resource gets deleted
	retry.Attempts(60)
	err = retry.Do(
		func() error {
			err := client.GetMembership(membershipID, true)
			if err != nil && client.Resource.State.Code == "DELETING" {
				return nil
			}
			return retry.Unrecoverable(err)
		})

	if err != nil {
		return fmt.Errorf("Waiting for resource to me deleted: %w", err)
	}

	// Delete K8s artifacts if deleteArtifacts is set to true
	if deleteArtifacts {
		err = k8s.DeleteArtifacts(ctx, k8sAuth)
		if err != nil {
			return fmt.Errorf("Deleting artifacts: %w", err)
		}
	}

	return nil
}

// ConnectAgent holds info needed to request and process a gek-connect-agent object
type ConnectAgent struct {
	Proxy                  string
	Namespace              string
	Version                string
	IsUpgrade              bool
	Registry               string
	ImagePullSecretContent string
	Response               ConnectManifestResponse
}

// InstallOrUpdateConnectAgent retrieves the connect-agent manifests from the gke api
// and installs or update them into a Kubernetes cluster
func (ca ConnectAgent) InstallOrUpdateConnectAgent(project string, membershipID string, k8sAuth k8s.Auth) error {
	client, err := NewClient(ctx, project, k8sAuth)
	if err != nil {
		return fmt.Errorf("Getting new membership client: %w", err)
	}

	// Get membership info
	err = client.GetMembership(membershipID, false)
	if err != nil {
		return fmt.Errorf("Checking membership info: %w", err)
	}

	// Call the api and get the manifests
	ca.Response, err = client.GenerateConnectManifest(ca.Proxy, ca.Namespace, ca.Version, ca.IsUpgrade, ca.Registry, ca.ImagePullSecretContent)
	if err != nil {
		return fmt.Errorf("Generating connect-agent manifests: %w", err)
	}

	debug.GoLog("InstallOrUpdateConnectAgent: first response manifest: " + ca.Response.Manifest[0].Manifest)
	return nil
}
