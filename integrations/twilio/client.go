package twilio

import (
	"context"

	"github.com/twilio/twilio-go"

	"go.autokitteh.dev/autokitteh/integrations/twilio/webhooks"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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

	getVars := func(ctx context.Context, cid sdktypes.ConnectionID) (*webhooks.Vars, error) {
		vs, err := vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return nil, err
		}

		var data webhooks.Vars
		vs.Decode(&data)
		return &data, nil
	}

	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(
			sdkmodule.ExportFunction("create_message",
				i.createMessage,
				sdkmodule.WithFuncDoc("https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource"),
				sdkmodule.WithArgs("to", "from_number?", "messaging_service_sid?", "body?", "media_url?", "content_sid?"),
			),
		),
		sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
			data, err := getVars(ctx, cid)
			if err != nil {
				return sdktypes.NewErrorStatus(err), nil
			}

			if data.AccountSID == "" {
				return sdktypes.NewErrorStatus(sdkerrors.ErrNotInitialized), nil
			}

			return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
		}),
		sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
			data, err := getVars(ctx, cid)
			if err != nil {
				return sdktypes.NewErrorStatus(err), nil
			}

			client := twilio.NewRestClientWithParams(twilio.ClientParams{
				AccountSid: data.AccountSID,
				Username:   data.Username,
				Password:   data.Password,
			})

			acct, err := client.Api.FetchAccount(data.AccountSID)
			if err != nil {
				return sdktypes.NewErrorStatus(err), nil
			}

			return sdktypes.NewStatusf(sdktypes.StatusCodeOK, "account %v: %v", acct.FriendlyName, acct.Status), nil
		}),
	)
}
