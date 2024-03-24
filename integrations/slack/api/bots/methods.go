package bots

import (
	"context"
	"errors"
	"net/url"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type API struct {
	Secrets sdkservices.Secrets
	Scope   string
}

// Info gets information about a bot user.
// Based on: https://api.slack.com/methods/bots.info.
// Required Slack app scopes: https://api.slack.com/scopes/users:read.
func (a API) Info(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		bot, teamID string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"bot", &bot,
		"team_id?", &teamID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	req := url.Values{}
	req.Set("bot", bot)
	if teamID != "" {
		req.Set("team_id", teamID)
	}

	// Invoke the API method.
	// TODO: Use HTTP GET instead of POST.
	resp := &InfoResponse{}
	err = api.PostForm(ctx, a.Secrets, a.Scope, req, resp, "bots.info")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// Info is only used internally, to check the usability of a bot token.
func InfoWithToken(ctx context.Context, secrets sdkservices.Secrets, scope, botToken string, authTest *auth.TestResponse) (*InfoResponse, error) {
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey{}, botToken)
	req := url.Values{}
	req.Set("bot", authTest.BotID)
	req.Set("team_id", authTest.TeamID)
	resp := &InfoResponse{}
	err := api.PostForm(ctx, secrets, scope, req, resp, "bots.info")
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	if resp.Bot.AppID == "" {
		return nil, errors.New("empty response")
	}
	return resp, nil
}
