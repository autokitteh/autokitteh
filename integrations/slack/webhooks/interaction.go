package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/chat"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/conversations"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/users"
)

const (
	InteractionPath = "/slack/interaction"
)

// https://api.slack.com/reference/interaction-payloads/block-actions#fields
type BlockActionsPayload struct {
	// Type must be "block_actions" (or "interactive_message" for attachments,
	// which we don't support because they're superseded by blocks).
	Type      string      `json:"type,omitempty"`
	User      *users.User `json:"user,omitempty"`
	Container *Container  `json:"container,omitempty"`
	// Contains data from the specific interactive component that was used.
	// App surfaces can contain blocks with multiple interactive components,
	// and each of those components can have multiple values selected by users.
	Actions []Action `json:"actions,omitempty"`
	// TODO: "state": {"values":{}}

	APIAppID            string      `json:"api_app_id,omitempty"`
	Team                *Team       `json:"team,omitempty"`
	IsEnterpriseInstall bool        `json:"is_enterprise_install,omitempty"`
	Enterprise          *Enterprise `json:"enterprise,omitempty"`

	Channel *conversations.Channel `json:"channel,omitempty"`
	Message *chat.Message          `json:"message,omitempty"`

	// Short-lived webhook URL to send messages in response to interactions.
	// Attention: documented as deprecated for next-generation Slack apps, see
	// https://api.slack.com/reference/interaction-payloads/block-actions#fields
	// and compare with [webhooks.SlashCommand].
	ResponseURL string `json:"response_url,omitempty"`
	// Short-lived ID that will let your app open a modal
	// (https://api.slack.com/surfaces/modals).
	TriggerID string `json:"trigger_id,omitempty"`
}

// https://api.slack.com/reference/interaction-payloads/block-actions
type Action struct {
	Type string `json:"type,omitempty"`
	// Identifies the block within a surface that contained the interactive
	// component that was used. See https://api.slack.com/reference/block-kit/block-elements.
	BlockID  string `json:"block_id,omitempty"`
	ActionTS string `json:"action_ts,omitempty"`

	// Identifies the interactive component itself. Some blocks can contain
	// multiple interactive components, so [BlockID] alone may not be specific
	// enough to identify the source component. For more information, see
	// https://api.slack.com/reference/block-kit/block-elements.
	ActionID string `json:"action_id,omitempty"`
	// Set by your app when you composed the blocks, this is the value that was
	// specified in the interactive component when an interaction happened. For
	// example, a select menu will have multiple possible values depending on
	// what the user picks from the menu, and Value will identify the chosen
	// option. See https://api.slack.com/reference/block-kit/block-elements.
	Value string `json:"value,omitempty"`

	Style string     `json:"style,omitempty"`
	Text  *chat.Text `json:"text,omitempty"`
}

// https://api.slack.com/reference/interaction-payloads/block-actions#examples
type Container struct {
	Type        string `json:"type,omitempty"`
	MessageTS   string `json:"message_ts,omitempty"`
	ChannelID   string `json:"channel_id,omitempty"`
	IsEphemeral bool   `json:"is_ephemeral,omitempty"`
	ViewID      string `json:"view_id,omitempty"`
}

// https://api.slack.com/methods/oauth.v2.access#examples
type Enterprise struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// https://api.slack.com/methods/oauth.v2.access#examples
// https://api.slack.com/reference/interaction-payloads/block-actions#examples
type Team struct {
	ID     string `json:"id,omitempty"`
	Domain string `json:"domain,omitempty"`
	Name   string `json:"name,omitempty"`
}

type Response struct {
	Text            string           `json:"text,omitempty"`
	Blocks          []map[string]any `json:"blocks,omitempty"`
	ResponseType    string           `json:"response_type,omitempty"`
	ThreadTS        string           `json:"thread_ts,omitempty"`
	ReplaceOriginal bool             `json:"replace_original,omitempty"`
	DeleteOriginal  bool             `json:"delete_original,omitempty"`
}

// HandleInteraction dispatches and acknowledges a user interaction callback
// from Slack, e.g. shortcuts, interactive components in messages and modals,
// and Slack workflow steps. See https://api.slack.com/messaging/interactivity
// and https://api.slack.com/interactivity/handling. Compare this function
// with the [websockets.HandleInteractiveEvent] implementation.
func (h handler) HandleInteraction(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", InteractionPath))

	// Validate and parse the inbound request.
	body := checkRequest(w, r, l, api.ContentTypeForm)
	if body == nil {
		return
	}

	j, err := url.QueryUnescape(string(body))
	if err != nil {
		l.Error("Failed to URL-decode interaction callback",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	j = strings.TrimPrefix(j, "payload=")
	payload := &BlockActionsPayload{}
	if err := json.Unmarshal([]byte(j), payload); err != nil {
		l.Error("Failed to parse URL-decoded JSON payload",
			zap.String("json", j),
			zap.Error(err),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(l, payload, "interaction")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	enterpriseID := ""
	if payload.IsEnterpriseInstall {
		enterpriseID = payload.User.EnterpriseUser.EnterpriseID
	}
	cids, err := h.listConnectionIDs(ctx, payload.APIAppID, enterpriseID, payload.Team.ID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)

	// It's a Slack best practice to update an interactive message after the interaction,
	// to prevent further interaction with the same message, and to reflect the user actions.
	// See: https://api.slack.com/interactivity/handling#updating_message_response.
	h.updateMessage(ctx, payload)
}

// updateMessage updates an interactive message after the interaction, to prevent
// further interaction with the same message, and to reflect the user actions.
// See: https://api.slack.com/interactivity/handling#updating_message_response.
func (h handler) updateMessage(ctx context.Context, payload *BlockActionsPayload) {
	resp := Response{
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
	meta := &chat.UpdateResponse{}
	err := api.PostJSON(ctx, h.vars, resp, meta, payload.ResponseURL)
	if err != nil {
		l := extrazap.ExtractLoggerFromContext(ctx)
		l.Warn("Error in reply to user via interaction webhook",
			zap.Error(err),
			zap.String("url", payload.ResponseURL),
			zap.Any("response", resp),
		)
	}
}
