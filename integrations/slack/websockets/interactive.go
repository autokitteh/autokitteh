package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"html"

	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/chat"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// HandleInteraction dispatches and acknowledges a user interaction callback
// from Slack, e.g. shortcuts, interactive components in messages and modals,
// and Slack workflow steps. See https://api.slack.com/messaging/interactivity
// and https://api.slack.com/interactivity/handling. Compare this function
// with the [webhooks.HandleInteraction] implementation.
func (h handler) handleInteractiveEvent(e *socketmode.Event, c *socketmode.Client) {
	defer c.Ack(*e.Request)

	// Reuse the Slack event's JSON payload instead of the struct.
	body, err := e.Request.Payload.MarshalJSON()
	if err != nil {
		h.logger.Error("Bad request from Slack websocket",
			zap.Any("payload", e.Request.Payload),
		)
		return
	}

	// Parse the inbound request (no need to validate authenticity, unlike webhooks).
	payload := &webhooks.BlockActionsPayload{}
	if err := json.Unmarshal(body, payload); err != nil {
		h.logger.Error("Failed to parse interactive event's JSON payload",
			zap.ByteString("json", body),
			zap.Error(err),
		)
		return
	}

	// Transform the received Slack event into an autokitteh event.
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(payload)
	if err != nil {
		h.logger.Error("Failed to wrap Slack event",
			zap.Any("payload", payload),
			zap.Error(err),
		)
		return
	}

	m, err := wrapped.ToStringValuesMap()
	if err != nil {
		h.logger.Error("Failed to convert wrapped Slack event",
			zap.Any("payload", payload),
			zap.Error(err),
		)
		return
	}

	pb := kittehs.TransformMapValues(m, sdktypes.ToProto)
	akEvent := &sdktypes.EventPB{
		EventType: "interaction",
		Data:      pb,
	}

	// Retrieve all the relevant connections for this event.
	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.WebSocketName, "")
	if err != nil {
		h.logger.Error("Failed to retrieve connection tokens", zap.Error(err))
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(cids, akEvent)

	// It's a Slack best practice to update an interactive message after the interaction,
	// to prevent further interaction with the same message, and to reflect the user actions.
	// See: https://api.slack.com/interactivity/handling#updating_message_response.
	h.updateMessage(payload, cids)
}

// updateMessage updates an interactive message after the interaction, to prevent
// further interaction with the same message, and to reflect the user actions.
// See: https://api.slack.com/interactivity/handling#updating_message_response.
func (h handler) updateMessage(payload *webhooks.BlockActionsPayload, cids []sdktypes.ConnectionID) {
	resp := webhooks.Response{
		Text:            payload.Message.Text,
		ResponseType:    "in_channel",
		ReplaceOriginal: true,
	}
	if payload.Container.IsEphemeral {
		resp.ResponseType = "ephemeral"
	}

	// Copy all the message's blocks, except actions.
	for _, b := range payload.Message.Blocks {
		if b.Type == "header" {
			b.Text.Text = html.UnescapeString(b.Text.Text)
		}
		if b.Type != "actions" {
			resp.Blocks = append(resp.Blocks, b)
		}
	}

	// And append new blocks to reflect the user actions.
	for _, a := range payload.Actions {
		if a.Type == "button" {
			action := "<@%s> clicked the `%s` button"
			action = fmt.Sprintf(action, payload.User.ID, payload.Actions[0].Text.Text)
			switch payload.Actions[0].Style {
			case "primary":
				action = ":large_green_square: " + action
			case "danger":
				action = ":large_red_square: " + action
			}
			resp.Blocks = append(resp.Blocks, chat.Block{
				Type: "section",
				Text: &chat.Text{
					Type: "mrkdwn",
					Text: action,
				},
			})
		}
	}

	// Send the update to Slack's webhook.
	appToken := h.firstBotToken(cids)
	meta := &chat.UpdateResponse{}
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey{}, appToken)
	err := api.PostJSON(ctx, h.vars, resp, meta, payload.ResponseURL)
	if err != nil {
		h.logger.Warn("Error in reply to user via interaction webhook",
			zap.String("url", payload.ResponseURL),
			zap.Any("response", resp),
			zap.Error(err),
		)
	}
}

// Return the Slack bot token of the first connection, if there is any.
func (h handler) firstBotToken(cids []sdktypes.ConnectionID) string {
	for _, cid := range cids {
		if data, err := h.vars.Get(context.Background(), sdktypes.NewVarScopeID(cid), vars.BotTokenName); err == nil {
			return data.GetValue(vars.BotTokenName)
		}
	}
	// This will result in a warning in the server's log,
	// but the caller (updateMessage()) should still work.
	return ""
}
