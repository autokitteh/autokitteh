// Package bots is a lightweight wrapper over the "bots" methods
// in Slack's Web API: https://api.slack.com/methods?filter=bots.
package bots

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/bots.info#examples
type InfoResponse struct {
	api.SlackResponse

	Bot Bot `json:"bot,omitempty"`
}

// -------------------- Auxiliary data structures --------------------

// https://api.slack.com/methods/bots.info#examples
type Bot struct {
	ID      string            `json:"id,omitempty"`
	Deleted bool              `json:"deleted,omitempty"`
	Name    string            `json:"name,omitempty"`
	Updated int               `json:"updated,omitempty"`
	AppID   string            `json:"app_id,omitempty"`
	UserID  string            `json:"user_id,omitempty"`
	Icons   map[string]string `json:"icons,omitempty"`
}
