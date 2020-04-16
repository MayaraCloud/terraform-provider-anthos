package main

import (
        "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// Provider returns the map of Terraform resources
func Provider() *schema.Provider {
        return &schema.Provider{
                ResourcesMap: map[string]*schema.Resource{
                        "anthos_cluster_membership": resourceMembership(),
                },
        }
}