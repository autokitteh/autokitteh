// Package auth is a lightweight wrapper over the "auth" methods
// in Slack's Web API: https://api.slack.com/methods?filter=auth.
package auth

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/auth.test#examples
type TestResponse struct {
	api.SlackResponse

	URL                 string `json:"url,omitempty"`
	Team                string `json:"team,omitempty"`
	User                string `json:"user,omitempty"`
	TeamID              string `json:"team_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
	BotID               string `json:"bot_id,omitempty"`
	EnterpriseID        string `json:"enterprise_id,omitempty"`
	IsEnterpriseInstall bool   `json:"is_enterprise_install,omitempty"`
}
