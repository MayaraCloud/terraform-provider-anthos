package main

import (
	"fmt"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceRegisterCluster() *schema.Resource {
    return &schema.Resource{
        Create: resourceRegisterClusterCreate,
        Read:   resourceRegisterClusterRead,
        Update: resourceRegisterClusterUpdate,
        Delete: resourceRegisterClusterDelete,

        Schema: map[string]*schema.Schema{
            "testfield": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },
        },
    }
}

func resourceRegisterClusterCreate(d *schema.ResourceData, m interface{}) error {
	return resourceRegisterClusterRead(d, m)
}

func resourceRegisterClusterRead(d *schema.ResourceData, m interface{}) error {
	fmt.Println("Testing Terraform Anthos provider")
	return nil
}

func resourceRegisterClusterUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceRegisterClusterRead(d, m)
}

func resourceRegisterClusterDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
