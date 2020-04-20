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
    fmt.Println(hub.GetMembership("mayara-anthos", "mayara-gke", "", "", "", "", ))
}