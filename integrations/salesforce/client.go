package salesforce

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	integrationID = sdktypes.NewIntegrationIDFromName("salesforce")
	AuthType      = sdktypes.NewSymbol("authType")
	OAuthDataName = sdktypes.NewSymbol("OAuthData")
)

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "salesforce",
	DisplayName:   "Salesforce",
	Description:   "Salesforce is a cloud-based customer relationship management (CRM) platform.",
	LogoUrl:       "/static/images/salesforce.png",
	UserLinks: map[string]string{
		"1 Salesforce API reference": "https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/intro.htm",
		"2 Python client API":        "https://github.com/simple-salesforce/simple-salesforce",
	},
	ConnectionUrl: "/salesforce/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(vs sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New( /* No Starlark functions */ ),
		connStatus(vs),
		connTest(vs),
	)
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Using X".
func connStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth v2"), nil
	})
}

// connTest is an optional connection test provided by the
// integration to AutoKitteh.
func connTest(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		// TODO: Implement connection test.

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}
