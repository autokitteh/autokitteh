package zoom

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	integrationName = "zoom"
)

var (
	integrationID = sdktypes.NewIntegrationIDFromName(integrationName)
	desc          = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
		IntegrationId: integrationID.String(),
		UniqueName:    integrationName,
		DisplayName:   "Zoom",
		LogoUrl:       "/static/images/zoom.svg",
		ConnectionUrl: "/zoom",
		ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
			RequiresConnectionInit: true,
		},
	}))
)

// New defines an AutoKitteh integration, which
// is registered when the AutoKitteh server starts.
func New(v sdkservices.Vars, o sdkservices.OAuth) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		sdkintegrations.WithConnectionConfigFromVars(v))
}
