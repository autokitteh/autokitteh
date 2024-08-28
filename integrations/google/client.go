package google

import (
	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/sheets"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	gmailIntegrationID = sdktypes.NewIntegrationIDFromName("gmail")
	formsIntegrationID = sdktypes.NewIntegrationIDFromName("googleforms")
	googleIntegrationID = sdktypes.NewIntegrationIDFromName("google")
)

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: googleIntegrationID.String(),
	UniqueName:    "google",
	DisplayName:   "Google (All APIs)",
	Description:   "Aggregation of all available Google APIs.",
	LogoUrl:       "/static/images/google.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://developers.google.com/apis-explorer",
		"2 Go client API":      "https://pkg.go.dev/google.golang.org/api",
		"3 Python samples":     "https://github.com/googleworkspace/python-samples",
	},
	ConnectionUrl: "/google/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	scope := desc.UniqueName().String()

	opts := gmail.ExportedFunctions(cvars, scope, true)
	opts = append(opts, sheets.ExportedFunctions(cvars, scope, true)...)

	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(opts...),
		connections.ConnStatus(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars))
}
