package drive

import (
	"go.autokitteh.dev/autokitteh/integrations/google/connections"
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
	LogoUrl:       "/static/images/google_drive.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://developers.google.com/drive/api/reference/rest/v3",
		"2 Python client API":  "https://developers.google.com/resources/api-libraries/documentation/drive/v3/python/latest/",
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
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}
