package twilio

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/twilio/webhooks"
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
	i := &integration{vars: vars}

	return sdkintegrations.NewIntegration(
		desc,
		sdkmodule.New(
			sdkmodule.ExportFunction("create_message",
				i.createMessage,
				sdkmodule.WithFuncDoc("https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource"),
				sdkmodule.WithArgs("to", "from_number?", "messaging_service_sid?", "body?", "media_url?", "content_sid?"),
			),
		),
		connStatus(i),
		connTest(i),
		sdkintegrations.WithConnectionConfigFromVars(vars),
	)
}

// connStatus is an optional connection status check provided by
// the integration to AutoKitteh. The possible results are "Init
// required" (the connection is not usable yet) and "Using X".
func connStatus(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionStatus(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		var decodedVars webhooks.Vars
		vs.Decode(&decodedVars)

		at := vs.Get(webhooks.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		switch at.Value() {
		case integrations.APIKey:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using API key"), nil
		case integrations.APIToken:
			return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using auth token"), nil
		default:
			return sdktypes.NewStatus(sdktypes.StatusCodeError, "Bad auth type"), nil
		}
	})
}

// connTest is an optional connection test provided by the integration
// to AutoKitteh. It is used to verify that the connection is working
// as expected. The possible results are "OK" and "error".
func connTest(i *integration) sdkintegrations.OptFn {
	return sdkintegrations.WithConnectionTest(func(ctx context.Context, cid sdktypes.ConnectionID) (sdktypes.Status, error) {
		if !cid.IsValid() {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		vs, err := i.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
		if err != nil {
			return sdktypes.InvalidStatus, err
		}

		at := vs.Get(webhooks.AuthType)
		if !at.IsValid() || at.Value() == "" {
			return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
		}

		var decodedVars webhooks.Vars
		vs.Decode(&decodedVars)
		accountSID := decodedVars.AccountSID
		authSID := accountSID
		authToken := ""

		if at.Value() == integrations.APIKey {
			authSID = decodedVars.Username
		}
		authToken = decodedVars.Password

		url := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s.json", accountSID)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		req.SetBasicAuth(authSID, authToken)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}
		defer resp.Body.Close()

		_, err = io.ReadAll(resp.Body)
		if err != nil {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, err.Error()), nil
		}

		if resp.StatusCode != http.StatusOK {
			return sdktypes.NewStatus(sdktypes.StatusCodeError, fmt.Sprintf("Request failed. Status Code: %d", resp.StatusCode)), nil
		}

		return sdktypes.NewStatus(sdktypes.StatusCodeOK, ""), nil
	})
}
