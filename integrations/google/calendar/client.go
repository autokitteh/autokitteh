package calendar

import (
	"context"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	googleScope = "google"
)

type api struct {
	Vars  sdkservices.Vars
	Scope string
}

var integrationID = sdktypes.NewIntegrationIDFromName("googlecalendar")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "googlecalendar",
	DisplayName:   "Google Calendar",
	Description:   "Google Calendar is a time-management and scheduling calendar service developed by Google.",
	LogoUrl:       "/static/images/google_calendar.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://developers.google.com/calendar/api/v3/reference",
		"2 Go client API":      "https://pkg.go.dev/google.golang.org/api/calendar/v3",
		"3 Python client API":  "https://developers.google.com/resources/api-libraries/documentation/calendar/v3/python/latest/",
		"4 Python samples":     "https://github.com/googleworkspace/python-samples/tree/main/calendar",
	},
	ConnectionUrl: "/googlecalendar/connect",
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
