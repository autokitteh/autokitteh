package api

import (
	"context"
	"errors"
	"fmt"
)

// https://api.slack.com/methods/bots.info#examples
type BotsInfoResponse struct {
	SlackResponse

	Bot Bot `json:"bot"`
}

// https://api.slack.com/methods/bots.info#examples
type Bot struct {
	ID      string            `json:"id"`
	Deleted bool              `json:"deleted"`
	Name    string            `json:"name"`
	Updated int               `json:"updated"`
	AppID   string            `json:"app_id"`
	UserID  string            `json:"user_id"`
	Icons   map[string]string `json:"icons"`
}

// BotsInfo gets information about a bot user.
// Based on: https://api.slack.com/methods/bots.info
// Required Slack app scope: https://api.slack.com/scopes/users:read
func BotsInfo(ctx context.Context, botToken string, authTest *AuthTestResponse) (*Bot, error) {
	req := fmt.Sprintf("bots.info?bot=%s&team_id=%s", authTest.BotID, authTest.TeamID)
	if authTest.IsEnterpriseInstall {
		req = fmt.Sprintf("%s&enterprise_id=%s", req, authTest.EnterpriseID)
	}

	resp := &BotsInfoResponse{}
	if err := get(ctx, botToken, req, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	if resp.Bot.AppID == "" {
		return nil, errors.New("empty response")
	}
	return &resp.Bot, nil
}
