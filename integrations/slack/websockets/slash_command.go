package websockets

import (
	"context"
	"fmt"
	"strings"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/api/chat"
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// HandleSlashCommand dispatches and acknowledges a user's slash command registered by our
// Slack app. See https://api.slack.com/interactivity/slash-commands#responding_to_commands.
// Compare this function with the [webhooks.HandleSlashCommand] implementation.
func (h handler) handleSlashCommand(e *socketmode.Event, c *socketmode.Client) {
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

	// Transform the slash command struct into an autokitteh event.
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(cmd)
	if err != nil {
		h.logger.Error("Failed to wrap Slack event",
			zap.Any("cmd", cmd),
			zap.Error(err),
		)
		c.Ack(*e.Request)
		return
	}

	m, err := wrapped.ToStringValuesMap()
	if err != nil {
		h.logger.Error("Failed to convert wrapped Slack event",
			zap.Any("cmd", cmd),
			zap.Error(err),
		)
		c.Ack(*e.Request)
		return
	}

	pb := kittehs.TransformMapValues(m, sdktypes.ToProto)
	akEvent := &sdktypes.EventPB{
		IntegrationId:   h.integrationID.String(),
		OriginalEventId: cmd.TriggerID,
		EventType:       "slash_command",
		Data:            pb,
	}

	// Retrieve all the relevant connections for this event.
	connTokens, err := h.secrets.List(context.Background(), h.scope, "websockets")
	if err != nil {
		h.logger.Error("Failed to retrieve connection tokens", zap.Error(err))
		c.Ack(*e.Request)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(connTokens, akEvent)

	// https://api.slack.com/apis/connections/socket#acknowledge
	if len(cmd.Text) == 0 {
		c.Ack(*e.Request)
		return
	}

	// https://api.slack.com/apis/connections/socket#command
	c.Ack(*e.Request, map[string][]*chat.Block{
		"blocks": {
			{
				Type: "section",
				Text: &chat.Text{
					Type: "mrkdwn",
					Text: fmt.Sprintf("Your command: `%s`", cmd.Text),
				},
			},
		},
	})
}

func appSecretName(appID, enterpriseID, teamID string) string {
	s := fmt.Sprintf("apps/%s/%s/%s", appID, enterpriseID, teamID)
	// Slack enterprise ID is allowed to be empty.
	return strings.ReplaceAll(s, "//", "/")
}
