package calendar

import (
	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.NewIntegrationIDFromName("googlecalendar")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "googlecalendar",
	DisplayName:   "Google Calendar",
	Description:   "Google Calendar is a time-management and scheduling calendar service developed by Google.",
	LogoUrl:       "/static/images/google_calendar.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://developers.google.com/calendar/api/v3/reference",
		"2 Python client API":  "https://developers.google.com/resources/api-libraries/documentation/calendar/v3/python/latest/",
		"3 Python samples":     "https://github.com/googleworkspace/python-samples/tree/main/calendar",
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
		connections.ConnStatus(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}
