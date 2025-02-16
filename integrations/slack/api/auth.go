package api

import (
	"context"
	"errors"
	"net/http"
)

// https://api.slack.com/methods/auth.test#examples
type AuthTestResponse struct {
	SlackResponse

	URL                 string `json:"url"`
	Team                string `json:"team"`
	User                string `json:"user"`
	TeamID              string `json:"team_id"`
	UserID              string `json:"user_id"`
	BotID               string `json:"bot_id"`
	EnterpriseID        string `json:"enterprise_id,omitempty"`
	IsEnterpriseInstall bool   `json:"is_enterprise_install"`
}

// AuthTest checks the caller's authentication & identity.
// Based on: https://api.slack.com/methods/auth.test (no scopes required).
func AuthTest(ctx context.Context, botToken string) (*AuthTestResponse, error) {
	resp := &AuthTestResponse{}
	if err := Post(ctx, botToken, "auth.test", http.NoBody, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}
