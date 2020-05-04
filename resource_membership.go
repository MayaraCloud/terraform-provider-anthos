package main

import (
	"gitlab.com/mayara/private/anthos/k8s"
    "gitlab.com/mayara/private/anthos/hub"
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
            "hub_project_id": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                Description: "GCP project id in which the cluster will be registered",
            },
            "cluster_name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                Description: "Kubernetes cluster to register, this will be the cluster name in the hub",
                ForceNew: true,
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
            "delete_artifacts_on_destroy": &schema.Schema{
                Type:     schema.TypeBool,
                Default: true,
                Required: false,
                Optional: true,
                Description: "If true, when deleting the cluster from the Hub, delete also the artifacts installed in the Kubernetes cluster",
            },
        },
    }
}

func resourceMembershipCreate(d *schema.ResourceData, m interface{}) error {
    var k8sAuth k8s.Auth

    k8sAuth.KubeConfigFile = d.Get("k8s_config_file").(string)
    k8sAuth.KubeContext = d.Get("k8s_context").(string)
    clusterUUID, err := hub.CreateMembership(d.Get("hub_project_id").(string), d.Get("cluster_name").(string), "", d.Get("description").(string), "", k8sAuth)
    if err != nil {
        return fmt.Errorf("Creating Membership: %w", err)
    }
    d.SetId(clusterUUID)
	return resourceMembershipRead(d, m)
}

func resourceMembershipRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceMembershipUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceMembershipRead(d, m)
}

func resourceMembershipDelete(d *schema.ResourceData, m interface{}) error {
    var k8sAuth k8s.Auth
    
    k8sAuth.KubeConfigFile = d.Get("k8s_config_file").(string)
    k8sAuth.KubeContext = d.Get("k8s_context").(string)
    err := hub.DeleteMembership(d.Get("hub_project_id").(string), d.Get("cluster_name").(string), "", d.Get("description").(string), "", k8sAuth, d.Get("delete_artifacts_on_destroy").(bool))
    if err != nil {
        return fmt.Errorf("Deleting Membership: %w", err)
    }
	return nil
}
