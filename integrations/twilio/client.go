package twilio

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type integration struct {
	secrets sdkservices.Secrets
	scope   string
}

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
}))

func New(sec sdkservices.Secrets) sdkservices.Integration {
	i := integration{secrets: sec, scope: desc.UniqueName().String()}

	return sdkintegrations.NewIntegration(desc, sdkmodule.New(
		sdkmodule.WithConfigAsData(),
		sdkmodule.ExportFunction("create_message",
			i.createMessage,
			sdkmodule.WithFuncDoc("https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource"),
			sdkmodule.WithArgs("to", "from_number?", "messaging_service_sid?", "body?", "media_url?", "content_sid?"),
		),
	))
}
