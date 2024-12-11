package hubspot

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct{ vars sdkservices.Vars }

var (
	apiKey        = sdktypes.NewSymbol("api_key")
	authType      = sdktypes.NewSymbol("authType")
	integrationID = sdktypes.NewIntegrationIDFromName("hubspot")
)

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "hubspot",
	DisplayName:   "HubSpot",
	Description:   "HubSpot is an AI-powered customer platform that provides software, integrations, and resources to support marketing, sales, and customer service.",
	LogoUrl:       "/static/images/hubspot.svg",
	UserLinks: map[string]string{
		"HubSpot developer platform": "https://developers.hubspot.com/",
	},
	ConnectionUrl: "/hubspot/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connStatus(i),
		connTest(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// TODO: implement
func connStatus(_ *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
	})
}

// TODO: implement
func connTest(_ *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "OK"), nil
	})
}
