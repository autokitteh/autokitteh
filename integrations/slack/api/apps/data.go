// Package apps is a lightweight wrapper over the "apps" methods
// in Slack's Web API: https://api.slack.com/methods?filter=apps.
package apps

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/apps.connections.open#examples
type ConnectionsOpenResponse struct {
	api.SlackResponse

	URL string `json:"url,omitempty"`
}
