package main

import (
	"gitlab.com/mayara/private/anthos/k8s"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGkeConnectAgent() *schema.Resource {
    return &schema.Resource{
        Create: resourceGkeConnectAgentCreate,
        Read:   resourceGkeConnectAgentRead,
        Update: resourceGkeConnectAgentUpdate,
        Delete: resourceGkeConnectAgentDelete,

        Schema: map[string]*schema.Schema{
            "hub_project_id": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                Description: "GCP project id to which the hub registered cluster belongs",
            },
            "cluster_name": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
                Description: "Kubernetes cluster name in the hub registry",
                ForceNew: true,
            },
            "description": &schema.Schema{
                Type:     schema.TypeString,
                Required: false,
                Optional: true,
                Description: "Description of the gke connect agent",
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

func resourceGkeConnectAgentCreate(d *schema.ResourceData, m interface{}) error {
    var k8sAuth k8s.Auth

    k8sAuth.KubeConfigFile = d.Get("k8s_config_file").(string)
    k8sAuth.KubeContext = d.Get("k8s_context").(string)
	return resourceGkeConnectAgentRead(d, m)
}

func resourceGkeConnectAgentRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceGkeConnectAgentUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceGkeConnectAgentRead(d, m)
}

func resourceGkeConnectAgentDelete(d *schema.ResourceData, m interface{}) error {
    var k8sAuth k8s.Auth
    
    k8sAuth.KubeConfigFile = d.Get("k8s_config_file").(string)
    k8sAuth.KubeContext = d.Get("k8s_context").(string)
	return nil
}
