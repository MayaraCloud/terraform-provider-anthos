package hub

import (
	"fmt"
	"context"
)

// ParentRef is the resource name of the parent collection of a membership.
type ParentRef string

// GetParentRef gets the resource name of the parent collection of a membership.
func GetParentRef (project string, location string) ParentRef {
	return ParentRef(fmt.Sprintf("projects/%v/locations/%v", project, location))
}

// Global variable used for various context purposes
var ctx = context.Background()


// GetMembership gets a Membership resource from the GKEHub API
func GetMembership(project string, membershipID string, description string, gkeClusterSelfLink string, externalID string, issuerURL string) error {
	client, err := NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("Getting new client: %w", err)
	}
	return client.GetMembership(ctx, membershipID)
}

// CreateMembership creates a membership GKEHub resource 
func CreateMembership(project string, membershipID string, description string, gkeClusterSelfLink string, externalID string, issuerURL string, membershipManifest string) error {
	client, err := NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("Getting new client: %w", err)
	}
	
	// Populate the membership resource fields with the parameters
	client.Resource.Name = membershipID
	client.Resource.Description = membershipID
	client.Resource.ExternalID = externalID
	client.Resource.Endpoint.GKECluster.ResourceLink = gkeClusterSelfLink
	client.K8S.CRManifest = membershipManifest
	return client.CreateMembership(ctx)
}