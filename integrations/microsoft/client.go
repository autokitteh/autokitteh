package microsoft

import (
	"go.autokitteh.dev/autokitteh/integrations/microsoft/connection"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	integrationID = sdktypes.NewIntegrationIDFromName("microsoft")

	desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
		IntegrationId: integrationID.String(),
		UniqueName:    "microsoft",
		DisplayName:   "Microsoft (All APIs)",
		LogoUrl:       "/static/images/microsoft.svg",
		ConnectionUrl: "/microsoft/connect",
		ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
			RequiresConnectionInit: true,
		},
	}))
)

// New defines an AutoKitteh integration, which
// is registered when the AutoKitteh server starts.
func New(v sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc, sdkmodule.New(),
		connection.Status(v), // TODO: connection.Test(v),
		sdkintegrations.WithConnectionConfigFromVars(v))
}
