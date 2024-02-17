// Package reactions is a lightweight wrapper over the "reactions" methods
// in Slack's Web API: https://api.slack.com/methods?filter=reactions.
package reactions

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/reactions.add#args
type AddRequest struct {
	// Channel where the message to add reaction to was posted. Required.
	Channel string `json:"channel"`
	// Name of the reaction (emoji). Required.
	Name string `json:"name"`
	// Timestamp of the message to add reaction to. Required.
	Timestamp string `json:"timestamp"`
}

// https://api.slack.com/methods/reactions.add#examples
type AddResponse struct {
	api.SlackResponse
}
