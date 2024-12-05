package webhooks

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

const (
	SlashCommandPath = "/slack/command"
)

// See https://api.slack.com/interactivity/slash-commands#app_command_handling
// and https://api.slack.com/types/event.
type SlashCommand struct {
	// Unique identifier of the Slack workspace where the event occurred.
	TeamID string
	// Human-readable name of the Slack workspace where the event occurred.
	TeamDomain string

	// Is the executing Slack workspace part of an Enterprise Grid?
	IsEnterpriseInstall bool
	EnterpriseID        string
	EnterpriseName      string

	// APIAppID is our Slack app's unique ID. Useful in case we point multiple
	// Slack apps to the same webhook URL, but want to treat them differently
	// (e.g. official vs. unofficial, breaking changes, and flavors).
	APIAppID  string
	ChannelID string
	// Human-readable name of the channel - don't rely on it.
	ChannelName string
	// ID of the user who triggered the command.
	// Use "<@value>" in messages to mention them.
	UserID string
	// Command must be "/ak" or "/autokitteh" in our Slack app.
	Command string
	// Text that the user typed after the command (e.g. "help").
	Text string

	// Short-lived webhook URL (https://api.slack.com/messaging/webhooks) to generate
	// message responses (https://api.slack.com/interactivity/handling#message_responses).
	// Compare with [api.BlockActionsInteractionPayload], where this field is deprecated
	// per https://api.slack.com/reference/interaction-payloads/block-actions#fields.
	ResponseURL string
	// Short-lived ID that will let your app open a modal
	// (https://api.slack.com/surfaces/modals).
	TriggerID string
}

// HandleSlashCommand dispatches and acknowledges a user's slash command registered by our
// Slack app. See https://api.slack.com/interactivity/slash-commands#responding_to_commands.
// Compare this function with the [websockets.HandleSlashCommand] implementation.
func (h handler) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", SlashCommandPath))

	// Validate and parse the inbound request.
	body := h.checkRequest(w, r, l, api.ContentTypeForm)
	if body == nil {
		return
	}

	kv, err := url.ParseQuery(string(body))
	if err != nil {
		l.Error("Failed to parse slash command's URL-encoded form",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// See https://api.slack.com/interactivity/slash-commands#app_command_handling
	// (the informational note under the parameters table).
	if kv.Get("ssl_check") != "" {
		return
	}

	isEnterprise, err := strconv.ParseBool(kv.Get("is_enterprise_install"))
	if err != nil {
		isEnterprise = false
	}
	cmd := SlashCommand{
		TeamID:     kv.Get("team_id"),
		TeamDomain: kv.Get("team_domain"),

		IsEnterpriseInstall: isEnterprise,
		EnterpriseID:        kv.Get("enterprise_id"),
		EnterpriseName:      kv.Get("enterprise_name"),

		APIAppID:    kv.Get("api_app_id"),
		ChannelID:   kv.Get("channel_id"),
		ChannelName: kv.Get("channel_name"),
		UserID:      kv.Get("user_id"),
		Command:     kv.Get("command"),
		Text:        kv.Get("text"),

		ResponseURL: kv.Get("response_url"),
		TriggerID:   kv.Get("trigger_id"),
	}

	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(l, cmd, "slash_command")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := r.Context()
	cids, err := h.listConnectionIDs(ctx, cmd.APIAppID, cmd.EnterpriseID, cmd.TeamID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)

	// https://api.slack.com/interactivity/slash-commands#responding_to_commands
	// https://api.slack.com/interactivity/slash-commands#responding_response_url
	// https://api.slack.com/interactivity/slash-commands#enabling-interactivity-with-slash-commands__best-practices
	w.Header().Add(api.HeaderContentType, api.ContentTypeJSONCharsetUTF8)
	resp := "{\"response_type\": \"ephemeral\", \"text\": \"Your command: `%s %s`\"}"
	fmt.Fprintf(w, resp, cmd.Command, cmd.Text)
}
