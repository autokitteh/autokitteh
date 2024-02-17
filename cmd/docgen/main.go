// DocGen is an internal tool that exports metadata about the autokitteh
// CLI tool's commands, and the server's integration APIs, as Docusaurus
// markdown files for autokitteh's documentation website.
//
// The integration APIs are also exported as Python module stubs,
// for Python and Starlark language-server IntelliSense.
package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/cmd/docgen/integrations"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1/integrationsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
)

const (
	outputDir = "gen"
)

func main() {
	exportCLICommands()

	exportIntegrations([]integrations.Generator{
		integrations.NewLogoGenerator(outputDir),
		integrations.NewMarkdownGenerator(outputDir),
		integrations.NewPythonGenerator(outputDir),
	})
}

func exportCLICommands() {
	// TODO(ENG-415): Generate MDX files for each CLI command.
}

func exportIntegrations(gs []integrations.Generator) {
	url := sdkclient.DefaultLocalURL

	// Initialize a connection to the autokitteh server.
	ctx := context.Background()
	client := integrationsv1connect.NewIntegrationsServiceClient(
		http.DefaultClient, url,
	)

	// Fetch all the integrations from it.
	req := connect.NewRequest(&integrationsv1.ListRequest{})
	resp, err := client.List(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	// Generate output files for each one.
	for n, i := range resp.Msg.Integrations {
		log.Printf("Integration: %s", i.DisplayName)
		for _, g := range gs {
			log.Printf("  - Generating: %s", g.Output())
			g.Generate(url, n+1, i)
		}
	}
}
