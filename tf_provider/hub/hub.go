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
	return client.GetMembership(ctx, membershipID, false)
}

// CreateMembership creates a membership GKEHub resource 
func CreateMembership(project string, membershipID string, description string, gkeClusterSelfLink string, externalID string, issuerURL string, membershipManifest string) error {
	client, err := NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("Getting new client: %w", err)
	}
	// Check if membership does not already exist
	err = client.GetMembership(ctx, membershipID, true)	
	if err != nil {
		return fmt.Errorf("Checking if membership does not exist: %w", err)
	}

	// Populate the membership resource fields with the parameters
	client.Resource.Name = membershipID
	client.Resource.Description = membershipID
	client.Resource.ExternalID = externalID
	client.Resource.Endpoint.GKECluster.ResourceLink = gkeClusterSelfLink
	client.K8S.CRManifest = membershipManifest

	// Create the membership
	err = client.CreateMembership(ctx)
	if err != nil {
		return fmt.Errorf("Creating membership membership: %w", err)
	}

	// Get membership info after creation, just to double check that all went fine
	err = client.GetMembership(ctx, membershipID, false)	
	if err != nil {
		return fmt.Errorf("Checking getting membership info after creation: %w", err)
	}

	return nil
}

// DeleteMembership deletes a membership GKEHub resource 
func DeleteMembership(project string, membershipID string, description string, gkeClusterSelfLink string, externalID string, issuerURL string, membershipManifest string) error {
	client, err := NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("Getting new client: %w", err)
	}
	// Get membership info
	err = client.GetMembership(ctx, membershipID, false)	
	if err != nil {
		return fmt.Errorf("Checking getting membership info: %w", err)
	}

	// Delete the membership
	err = client.DeleteMembership(ctx)
	if err != nil {
		return fmt.Errorf("Deleting membership: %w", err)
	}

	// Check if membership does not exist anymore
	err = client.GetMembership(ctx, membershipID, true)	
	if err != nil {
		return fmt.Errorf("Checking if membership does not exist: %w", err)
	}

	return nil
}