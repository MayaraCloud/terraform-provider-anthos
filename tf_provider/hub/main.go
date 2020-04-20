package hub

import (
	"fmt"
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)
func main() {}

// ParentRef is the resource name of the parent collection of a membership.
type ParentRef string

// GetParentRef gets the resource name of the parent collection of a membership.
func GetParentRef (project string, location string) ParentRef {
	return ParentRef(fmt.Sprintf("projects/%v/locations/%v", project, location))
}

// GetMembership gets a Membership resource from the GKE Hub API
func GetMembership(project string, membershipID string, description string, gkeClusterSelfLink string, externalID string, issuerURL string) (membership *Resource, err error) {
	ctx := context.Background()
	// Get default credentials https://godoc.org/golang.org/x/oauth2/google
	creds, err := google.FindDefaultCredentials(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("Getting credentials: %w", err)
	}
	// Create google api options with the generated credentials
	options := option.WithCredentials(creds)

	client, err := NewClient(ctx, project, options)
	return client.GetMembership(ctx, membershipID)
}