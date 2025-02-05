package microsoft

import (
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/connection"
	"go.autokitteh.dev/autokitteh/integrations/microsoft/teams"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	integrationName = "microsoft"
)

var desc = common.Descriptor(integrationName, "Microsoft (All APIs)", "/static/images/microsoft.svg")

// New defines an AutoKitteh integration, which
// is registered when the AutoKitteh server starts.
func New(v sdkservices.Vars, o sdkservices.OAuth) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc, sdkmodule.New(),
		connection.Status(v), connection.Test(v, o),
		sdkintegrations.WithConnectionConfigFromVars(v))
}

// resources returns the Microsoft Graph resources that each Microsoft integration
// should subscribe to in order to receive asynchronous change notifications.
func resources(i sdktypes.Integration) []string {
	// TODO: Convert this to a switch when we add more integrations.
	if i.UniqueName().String() == teams.IntegrationName {
		return teams.SubscriptionResources
	}
	return nil
}
