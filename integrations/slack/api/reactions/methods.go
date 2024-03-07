package reactions

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

type API struct {
	Secrets sdkservices.Secrets
	Scope   string
}

// Add a reaction (emoji) to an item (message).
//
// Based on: https://api.slack.com/methods/reactions.add
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/reactions:write
func (a API) Add(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var channel, name, timestamp string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &channel,
		"name", &name,
		"timestamp", &timestamp,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	req := AddRequest{
		Channel:   channel,
		Name:      name,
		Timestamp: timestamp,
	}

	// Invoke the API method.
	resp := &AddResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "reactions.add")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}
