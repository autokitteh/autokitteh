package google

import (
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/google/connections"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const IntegrationName = "google"

var desc = common.Descriptor(IntegrationName, "Google (All APIs)", "/static/images/google.svg")

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(),
		connections.ConnStatus(cvars),
		connections.ConnTest(cvars),
		sdkintegrations.WithConnectionConfigFromVars(cvars))
}
