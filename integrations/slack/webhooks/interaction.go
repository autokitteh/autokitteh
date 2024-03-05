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

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"

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
	Text            string       `json:"text,omitempty"`
	Blocks          []chat.Block `json:"blocks,omitempty"`
	ResponseType    string       `json:"response_type,omitempty"`
	ThreadTS        string       `json:"thread_ts,omitempty"`
	ReplaceOriginal bool         `json:"replace_original,omitempty"`
	DeleteOriginal  bool         `json:"delete_original,omitempty"`
}

// HandleInteraction dispatches and acknowledges a user interaction callback
// from Slack, e.g. shortcuts, interactive components in messages and modals,
// and Slack workflow steps. See https://api.slack.com/messaging/interactivity
// and https://api.slack.com/interactivity/handling.
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
			zap.Error(err),
			zap.ByteString("body", body),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	j = strings.TrimPrefix(j, "payload=")
	payload := &BlockActionsPayload{}
	if err := json.Unmarshal([]byte(j), payload); err != nil {
		l.Error("Failed to parse URL-decoded JSON payload",
			zap.Error(err),
			zap.String("json", j),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Transform the received Slack event into an autokitteh event.
	data, err := transformPayload(l, w, payload)
	if err != nil {
		return
	}
	akEvent := &sdktypes.EventPB{
		IntegrationId:   h.integrationID.String(),
		OriginalEventId: payload.TriggerID,
		EventType:       "interaction",
		Data:            data,
	}

	// Retrieve all the relevant connections for this event.
	enterpriseID := ""
	if payload.IsEnterpriseInstall {
		enterpriseID = payload.User.EnterpriseUser.EnterpriseID
	}
	connTokens, err := h.listTokens(payload.APIAppID, enterpriseID, payload.Team.ID)
	if err != nil {
		l.Error("Failed to retrieve connection tokens",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(l, connTokens, akEvent)

	// It's a Slack best practice to update an interactive message after the interaction, to
	// prevent further interaction with the same message, and to reflect the user actions. See:
	// https://api.slack.com/interactivity/handling#updating_message_response.
	h.updateMessage(l, payload)
}

// transformPayload transforms a received Slack event into an autokitteh event.
func transformPayload(l *zap.Logger, w http.ResponseWriter, payload *BlockActionsPayload) (map[string]*valuesv1.Value, error) {
	wrapped, err := sdkvalues.DefaultValueWrapper.Wrap(payload)
	if err != nil {
		l.Error("Failed to wrap Slack event",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Slack event",
			zap.Error(err),
			zap.Any("payload", payload),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	return kittehs.TransformMapValues(data, sdktypes.ToProto), nil
}

// updateMessage updates an interactive message after the interaction, to prevent
// further interaction with the same message, and to reflect the user actions. See:
// https://api.slack.com/interactivity/handling#updating_message_response.
func (h handler) updateMessage(l *zap.Logger, payload *BlockActionsPayload) {
	resp := Response{
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
	meta := &chat.UpdateResponse{}
	ctx := extrazap.AttachLoggerToContext(l, context.Background())
	err := api.PostJSON(ctx, h.secrets, h.scope, resp, meta, payload.ResponseURL)
	if err != nil {
		l.Warn("Error in reply to user via interaction webhook",
			zap.Error(err),
			zap.String("url", payload.ResponseURL),
			zap.Any("response", resp),
		)
	}
}
