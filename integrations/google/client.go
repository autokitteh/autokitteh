package google

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/sheets"
)

var integrationID = sdktypes.NewIntegrationIDFromName("google")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "google",
	DisplayName:   "Google (All APIs)",
	Description:   "Aggregation of all available Google APIs.",
	LogoUrl:       "/static/images/google.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://developers.google.com/apis-explorer",
		"2 Go client API":      "https://pkg.go.dev/google.golang.org/api",
	},
	ConnectionUrl: "/google/connect",
}))

func New(sec sdkservices.Secrets) sdkservices.Integration {
	scope := desc.UniqueName().String()
	opts := []sdkmodule.Optfn{sdkmodule.WithConfigAsData()}

	// TODO: Calendar.
	// TODO: Chat.
	// TODO: Drive.
	// TODO: Forms.
	opts = append(opts, gmail.ExportedFunctions(sec, scope, true)...)
	opts = append(opts, sheets.ExportedFunctions(sec, scope, true)...)

	return sdkintegrations.NewIntegration(desc, sdkmodule.New(opts...))
}
