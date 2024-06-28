package drive

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.NewIntegrationIDFromName("googledrive")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "googledrive",
	DisplayName:   "Google Drive",
	Description:   "Google Drive is a file-hosting service and synchronization service developed by Google.",
	LogoUrl:       "/static/images/google_forms.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://developers.google.com/drive/api/reference/rest/v3",
		"2 Python client API":  "https://developers.google.com/resources/api-libraries/documentation/drive/v3/python/latest/index.html",
		"3 Python samples":     "https://github.com/googleworkspace/python-samples/tree/main/drive",
	},
	ConnectionUrl: "/googledrive/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New( /* No exported functions for Starlark */ ),
		connStatus(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connStatus is an optional connection status check provided by the
// integration to AutoKitteh. The possible results are "init required"
// (the connection is not usable yet) and "using X" (where "X" is the
// authentication method: OAuth 2.0 (user), or JSON key (service account).
func connStatus(cvars sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := cvars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		if vs.Has(vars.OAuthData) {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using OAuth 2.0"), nil
		}
		if vs.Has(vars.JSON) {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using JSON key"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "unrecognized auth"), nil
	})
}
