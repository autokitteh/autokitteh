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
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/integrations/slack2/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleInteraction dispatches and acknowledges a user interaction callback
// from Slack, e.g. shortcuts, interactive components in messages and modals,
// and Slack workflow steps. See https://api.slack.com/messaging/interactivity
// and https://api.slack.com/interactivity/handling. Compare this function
// with the [webhooks.HandleInteraction] implementation.
func (h Handler) handleInteractiveEvent(e *socketmode.Event, c *socketmode.Client) {
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

	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(h.logger, payload, "interaction")
	if err != nil {
		return
	}

	// Retrieve all the relevant connections for this event.
	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.AppTokenVar, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
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
// See https://api.slack.com/interactivity/handling#updating_message_response.
func (h Handler) updateMessage(payload *webhooks.BlockActionsPayload, cids []sdktypes.ConnectionID) {
	resp := webhooks.Response{
		Text:            payload.Message.Text,
		ResponseType:    "in_channel",
		ReplaceOriginal: true,
	}
	if payload.Container.IsEphemeral {
		resp.ResponseType = "ephemeral"
	}

	// Copy all the message's blocks, except actions.
	// The event is verifiably from Slack, so we can trust the data.
	// TODO(ENG-1052): Support updating actions in non-last blocks.
	for _, b := range payload.Message.Blocks {
		// Header text is HTML-encoded, so unescape it.
		if b["type"] == "header" {
			h := b["text"].(map[string]any)
			h["text"] = html.UnescapeString(h["text"].(string))
		}
		if b["type"] != "actions" {
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
			resp.Blocks = append(resp.Blocks, map[string]any{
				"type": "section",
				"text": map[string]string{
					"type": "mrkdwn",
					"text": action,
				},
			})
		}
	}

	// Send the update to Slack's webhook.
	appToken := h.firstBotToken(cids)
	meta := &chat.UpdateResponse{}
	ctx := extrazap.AttachLoggerToContext(h.logger, context.Background())
	ctx = context.WithValue(ctx, api.OAuthTokenContextKey, appToken)
	err := api.PostJSON(ctx, h.vars, resp, meta, payload.ResponseURL)
	if err != nil {
		h.logger.Warn("Error in reply to user via interaction webhook",
			zap.String("url", payload.ResponseURL),
			zap.Any("response", resp),
			zap.Error(err),
		)
	}
}

// firstBotToken returns the Slack bot token of the first connection, if there is any.
func (h Handler) firstBotToken(cids []sdktypes.ConnectionID) string {
	for _, cid := range cids {
		if data, err := h.vars.Get(context.Background(), sdktypes.NewVarScopeID(cid), vars.BotTokenVar); err == nil {
			return data.GetValue(vars.BotTokenVar)
		}
	}
	// This will result in a warning in the server's log,
	// but the caller (updateMessage()) should still work.
	return ""
}
