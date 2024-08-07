package twilio

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct{ vars sdkservices.Vars }

var integrationID = sdktypes.NewIntegrationIDFromName("twilio")

var desc = kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
	IntegrationId: integrationID.String(),
	UniqueName:    "twilio",
	DisplayName:   "Twilio",
	Description:   "Twilio is a programmable phone-based communication platform: sending an receiving messages, making and receiving voice calls, and more.",
	LogoUrl:       "/static/images/twilio.png",
	UserLinks: map[string]string{
		"1 Messaging API overview": "https://www.twilio.com/docs/messaging/api",
		"2 Voice API overview":     "https://www.twilio.com/docs/voice/api",
	},
	ConnectionUrl: "/twilio/connect",
	ConnectionCapabilities: &sdktypes.ConnectionCapabilitiesPB{
		RequiresConnectionInit: true,
	},
}))

func New(vars sdkservices.Vars) sdkservices.Integration {
	i := integration{vars: vars}

	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(
			sdkmodule.ExportFunction("create_message",
				i.createMessage,
				sdkmodule.WithFuncDoc("https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource"),
				sdkmodule.WithArgs("to", "from_number?", "messaging_service_sid?", "body?", "media_url?", "content_sid?"),
			),
		),
		connStatus(vars),
		sdkintegrations.WithConnectionConfigFromVars(vars),
	)
}

func connStatus(vs sdkservices.Vars) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "init required"), nil
		}

		vs, err := vs.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		a := vs.Get(sdktypes.NewSymbol("AccountSID"))
		u := vs.Get(sdktypes.NewSymbol("Username"))
		if a.Name() == u.Name() {
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using auth token"), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, "using api key"), nil
	})
}
