package main

import (
    "gitlab.com/mayara/private/anthos/hub"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
    "fmt"
    "gitlab.com/mayara/private/anthos/k8s"
)

func resourceMembership() *schema.Resource {
    return &schema.Resource{
        Create: resourceMembershipCreate,
        Read:   resourceMembershipRead,
        Update: resourceMembershipUpdate,
        Delete: resourceMembershipDelete,

        Schema: map[string]*schema.Schema{
            "hub_project_id": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                Description: "GCP project id in which the cluster will be registered",
            },
            "cluster_name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                Description: "Kubernetes cluster to register, this will be the cluster name in the hub",
            },
            "description": &schema.Schema{
                Type:     schema.TypeString,
                Required: false,
                Optional: true,
                Description: "Description of the kubernetes cluster",
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
    kubeUUID, err := k8s.GetK8sClusterUUID(d)
    if err != nil {
        return fmt.Errorf("Getting uuid: %w", err)
    }
    d.SetId(kubeUUID)
    membershipManifest, err := k8s.GetMembershipCR(d)
    if err != nil {
        return fmt.Errorf("Getting membership k8s resource: %w", err)
    }
    err = hub.CreateMembership(d.Get("hub_project_id").(string), d.Get("cluster_name").(string), "", d.Get("description").(string), kubeUUID, "", membershipManifest)
    if err != nil {
        return fmt.Errorf("Creating Membership: %w", err)
    }
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
