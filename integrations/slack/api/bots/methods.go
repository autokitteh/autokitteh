package bots

import (
	"context"
	"errors"
	"net/url"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/auth"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type API struct {
	Vars sdkservices.Vars
}

// Info is only used internally, to check the usability of a bot token.
func InfoWithToken(ctx context.Context, botToken string, authTest *auth.TestResponse) (*InfoResponse, error) {
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey, botToken)
	req := url.Values{}
	req.Set("bot", authTest.BotID)
	req.Set("team_id", authTest.TeamID)
	resp := &InfoResponse{}
	err := api.PostForm(ctx, nil, req, resp, "bots.info")
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
