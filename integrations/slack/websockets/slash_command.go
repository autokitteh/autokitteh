package websockets

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/vars"
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
)

// handleSlashCommand dispatches and acknowledges a user's slash command registered by our
// Slack app. See https://api.slack.com/interactivity/slash-commands#responding_to_commands.
// Compare this function with the [webhooks.HandleSlashCommand] implementation.
func (h Handler) handleSlashCommand(e *socketmode.Event, c *socketmode.Client) {
	// This data casting is guaranteed to work, so no need to check it
	d := e.Data.(slack.SlashCommand)

	// Transform the received Slack event into an autokitteh slash command struct.
	cmd := webhooks.SlashCommand{
		TeamID:     d.TeamID,
		TeamDomain: d.TeamDomain,

		IsEnterpriseInstall: d.IsEnterpriseInstall,
		EnterpriseID:        d.EnterpriseID,
		EnterpriseName:      d.EnterpriseName,

		APIAppID:    d.APIAppID,
		ChannelID:   d.ChannelID,
		ChannelName: d.ChannelName,
		UserID:      d.UserID,
		Command:     d.Command,
		Text:        d.Text,

		ResponseURL: d.ResponseURL,
		TriggerID:   d.TriggerID,
	}

	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(h.logger, cmd, "slash_command")
	if err != nil {
		return
	}

	// Retrieve all the relevant connections for this event.
	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.AppTokenVar, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
		c.Ack(*e.Request)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(cids, akEvent)

	// https://api.slack.com/apis/connections/socket#command
	// https://api.slack.com/apis/connections/socket#acknowledge
	c.Ack(*e.Request, map[string][]map[string]any{
		"blocks": {
			{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": fmt.Sprintf("Your command: `%s %s`", cmd.Command, cmd.Text),
				},
			},
		},
	})
}
