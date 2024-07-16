package discord

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var integrationID = sdktypes.NewIntegrationIDFromName("discord")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "discord",
	DisplayName:   "Discord",
	Description:   "Discord is an instant messaging and VoIP social platform which allows communication through voice calls, video calls, text messaging, and media.",
	LogoUrl:       "/static/images/discord-icon.svg",
	UserLinks: map[string]string{
		"1 REST API reference": "https://discord.com/developers/docs/reference",
		"2 Python client API":  "https://discordpy.readthedocs.io/en/stable/api.html",
		"3 Python samples":     "https://github.com/Rapptz/discord.py/tree/master/example",
	},
	ConnectionUrl: "/discord/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(cvars sdkservices.Vars) sdkservices.Integration {
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New( /* No exported functions for Starlark */ ),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}
