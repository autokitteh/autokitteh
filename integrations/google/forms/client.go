package forms

import (
	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.NewIntegrationIDFromName("googleforms")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "googleforms",
	DisplayName:   "Google Forms",
	Description:   "Google Forms is a survey administration software that part of the Google Workspace office suite.",
	LogoUrl:       "/static/images/google_forms.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://developers.google.com/forms/api/reference/rest",
		"2 Python client API":  "https://googleapis.github.io/google-api-python-client/docs/dyn/forms_v1.html",
		"3 Python samples":     "https://github.com/googleworkspace/python-samples/tree/main/forms",
	},
	ConnectionUrl: "/googleforms/connect",
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
