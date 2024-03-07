package twilio

import (
	"context"
	"errors"

	"github.com/twilio/twilio-go"
	api "github.com/twilio/twilio-go/rest/api/v2010"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://www.twilio.com/docs/messaging/api/message-resource#create-a-message-resource
func (i integration) createMessage(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var to, from, messagingServiceSID, body, contentSID string
	var mediaURL []string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"to", &to,
		"from_number?", &from,
		"messaging_service_sid?", &messagingServiceSID,
		"body?", &body,
		"media_url?", &mediaURL,
		"content_sid?", &contentSID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	if from == "" && messagingServiceSID == "" {
		return sdktypes.InvalidValue, errors.New(`required: "from_number", or "messaging_service_sid"`)
	}
	if body == "" && len(mediaURL) == 0 && contentSID == "" {
		return sdktypes.InvalidValue, errors.New(`required: "body", "media_url", or "content_sid"`)
	}

	// Get auth details from secrets manager.
	token := sdkmodule.FunctionDataFromContext(ctx)
	auth, err := i.secrets.Get(context.Background(), i.scope, string(token))
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		AccountSid: auth["accountSID"],
		Username:   auth["username"],
		Password:   auth["password"],
	})
	params := &api.CreateMessageParams{}
	params.SetTo(to)
	if from != "" {
		params.SetFrom(from)
	}
	if messagingServiceSID != "" {
		params.SetMessagingServiceSid(messagingServiceSID)
	}
	if body != "" {
		params.SetBody(body)
	}
	if len(mediaURL) > 0 {
		params.SetMediaUrl(mediaURL)
	}
	if contentSID != "" {
		params.SetContentSid(contentSID)
	}
	// TODO: params.SetStatusCallback("https://.../twilio/xyz?")
	resp, err := client.Api.CreateMessage(params)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}
