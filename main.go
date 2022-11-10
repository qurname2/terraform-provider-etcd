package main

import (
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"terraform-provider-etcd/internal/provider"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: provider.New(),
		ProviderAddr: "qurname2/etcd",
		Debug:        debugMode,
	}

	logFlags := log.Flags()
	logFlags = logFlags &^ (log.Ldate | log.Ltime)
	log.SetFlags(logFlags)

	plugin.Serve(opts)
}
