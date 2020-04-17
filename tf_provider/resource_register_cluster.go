package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "fmt"
)

func resourceMembership() *schema.Resource {
    return &schema.Resource{
        Create: resourceMembershipCreate,
        Read:   resourceMembershipRead,
        Update: resourceMembershipUpdate,
        Delete: resourceMembershipDelete,

        Schema: map[string]*schema.Schema{
            "cluster_name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },
            "k8s_config_file": &schema.Schema{
                Type:     schema.TypeString,
                Required: false,
                Optional: true,
                Description: "Kubernetes specific credentials file",
                ConflictsWith: []string{"k8s_context"},
            },
            "k8s_context": &schema.Schema{
                Type:     schema.TypeString,
                Default: "current",
                Required: false,
                Optional: true,
                Description: "Use a context of the default credentials file",
                ConflictsWith: []string{"k8s_config_file"},
            },
        },
    }
}

func resourceMembershipCreate(d *schema.ResourceData, m interface{}) error {
    kubeUUID, err := getK8sClusterUUID(d)
    if err != nil {
        return fmt.Errorf("Getting uuid: %w", err)
    }
    d.SetId(kubeUUID)
	return resourceMembershipRead(d, m)
}

func resourceMembershipRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceMembershipUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceMembershipRead(d, m)
}

func resourceMembershipDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
