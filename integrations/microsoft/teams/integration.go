package teams

import (
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/connection"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const (
	IntegrationName = "microsoft_teams"
)

var desc = common.Descriptor(IntegrationName, "Microsoft Teams", "/static/images/microsoft_teams.svg").
	WithConnectionURL("/microsoft/teams")

// New defines an AutoKitteh integration, which
// is registered when the AutoKitteh server starts.
func New(v sdkservices.Vars, o sdkservices.OAuth) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc, sdkmodule.New(),
		connection.Status(v), connection.Test(v, o),
		sdkintegrations.WithConnectionConfigFromVars(v))
}
