package main

import (
	"fmt"
    "github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceMembership() *schema.Resource {
    return &schema.Resource{
        Create: resourceMembershipCreate,
        Read:   resourceMembershipRead,
        Update: resourceMembershipUpdate,
        Delete: resourceMembershipDelete,

        Schema: map[string]*schema.Schema{
            "testfield": &schema.Schema{
                Type:     schema.TypeString,
                Required: true,
            },
        },
    }
}

func resourceMembershipCreate(d *schema.ResourceData, m interface{}) error {
	return resourceMembershipRead(d, m)
}

func resourceMembershipRead(d *schema.ResourceData, m interface{}) error {
	fmt.Println("Testing Terraform Anthos provider")
	return nil
}

func resourceMembershipUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceMembershipRead(d, m)
}

func resourceMembershipDelete(d *schema.ResourceData, m interface{}) error {
	return nil
}
