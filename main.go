package main

import (
	"context"
	"flag"
	"log"

	"github.com/apollogeddon/ignition-tofu/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

//go:generate bash .github/scripts/ci/generate-docs.sh

// these will be set by the linker
var version = "1.2.4" // x-release-please-version

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/apollogeddon/ignition",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
