package main

import (
    "fmt"
	"github.com/MayaraCloud/terraform-provider-anthos/k8s"
	"github.com/MayaraCloud/terraform-provider-anthos/hub"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceGkeConnectAgent() *schema.Resource {
    return &schema.Resource{
        Create: resourceGkeConnectAgentCreate,
        Read:   resourceGkeConnectAgentRead,
        Update: resourceGkeConnectAgentUpdate,
        Delete: resourceGkeConnectAgentDelete,

        Schema: map[string]*schema.Schema{
            "project": &schema.Schema{
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
            "namespace": &schema.Schema{
                Type:     schema.TypeString,
                Default: "gke-connect",
                Required: false,
                Optional: true,
                Description: "Namespace to install connect agent to",
            },
            "proxy": &schema.Schema{
                Type:     schema.TypeString,
                Default: "",
                Required: false,
                Optional: true,
                Description: "URI of the proxy to reach gke-connect.googleapis.com.\nThe format must be in the form http(s)://{proxy_address},\ndepends on HTTP/HTTPS protocol supported by the proxy. This will direct\nconnect agent's outbound traffic through a HTTP(S) proxy",
            },
            "version": &schema.Schema{
                Type:     schema.TypeString,
                Default: "",
                Required: false,
                Optional: true,
                Description: "The version to use for connect agent.\nIf empty, the current default version will be use",
            },
            "is_upgrade": &schema.Schema{
                Type:     schema.TypeBool,
                Default: false,
                Required: false,
                Optional: true,
                Description: "If true, generate the resources for upgrade only. Some resources\n(e.g. secrets) generated for installation will be excluded",
            },
            "registry": &schema.Schema{
                Type:     schema.TypeString,
                Default: "",
                Required: false,
                Optional: true,
                Description: "The registry to fetch connect agent image; default to gcr.io/gkeconnect",
            },
            "image_pull_secret_content": &schema.Schema{
                Type:     schema.TypeString,
                Default: "",
                Required: false,
                Optional: true,
                Description: "The image pull secret content for the registry, if not public",
            },
        },
    }
}

func resourceGkeConnectAgentCreate(d *schema.ResourceData, m interface{}) error {
    var k8sAuth k8s.Auth
    k8sAuth.KubeConfigFile = d.Get("k8s_config_file").(string)
    k8sAuth.KubeContext = d.Get("k8s_context").(string)
    ca := initConnectAgent(d, m)
    err := ca.InstallOrUpdateConnectAgent(d.Get("project").(string), d.Get("cluster_name").(string), k8sAuth)
    if err != nil {
        return fmt.Errorf("Installing or updating connect agent: %w", err)
    }
    d.SetId("test")
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

func initConnectAgent(d *schema.ResourceData, m interface{}) hub.ConnectAgent {

    return hub.ConnectAgent{
    Proxy: d.Get("proxy").(string),
	Namespace: d.Get("namespace").(string),
	Version: d.Get("version").(string),
	IsUpgrade: d.Get("is_upgrade").(bool),
	Registry: d.Get("registry").(string),
    ImagePullSecretContent: d.Get("image_pull_secret_content").(string),
    }
}