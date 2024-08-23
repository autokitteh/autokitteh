package discord

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct{ vars sdkservices.Vars }

var (
	authType      = sdktypes.NewSymbol("authType")
	botToken      = sdktypes.NewSymbol("BotToken")
	integrationID = sdktypes.NewIntegrationIDFromName("discord")
)

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "discord",
	DisplayName:   "Discord",
	Description:   "Discord is an instant messaging and VoIP social platform which allows communication through voice calls, video calls, text messaging, and media.",
	LogoUrl:       "/static/images/discord.svg",
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
	i := &integration{vars: cvars}
	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New( /* No exported functions for Starlark */ ),
		connStatus(i),
		sdkintegrations.WithConnectionConfigFromVars(cvars),
	)
}

// connStatus is an optional connection status check provided by the
// integration with AutoKitteh. The possible results are "init required"
// (indicating the connection is not yet usable), "using X" (indicating
// one of multiple available authentication methods is in use), or
// "initialized" when only one authentication method is available.
func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(authType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		// Align with:
		// https://github.com/autokitteh/web-platform/blob/main/src/enums/connections/connectionTypes.enum.ts
		switch at.Value() {
		case "botToken":
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "initialized"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "bad auth type"), nil
		}
	})
}
