package main

import (
    "fmt"
    "gitlab.com/mayara/private/anthos/hub"
    "github.com/hashicorp/terraform-plugin-sdk/plugin"
    "github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// DebugMode is used only for development purposes, if set to true, then debug() gets executed instead of main()
const DebugMode = true

func main() {
    if DebugMode {
        debug()
    } else {
        plugin.Serve(&plugin.ServeOpts{
            ProviderFunc: func() terraform.ResourceProvider {
                return Provider()
            },
        })
    }
}

func debug() {
    client, _ := hub.GetMembership("mayara-anthos", "mayara-gke", "", "", "", "", )
    fmt.Println("Resource: ", client.Resource)
    fmt.Println("Calling update function:")
    hub.CreateMembership("mayara-anthos", "mayara-fake", "fake_description", "", "814a82ce-80da-4ad9-b6b4-30aaaaaa7777", "")
}