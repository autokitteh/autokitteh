package webhooks

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

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
func (h handler) HandleSlashCommand(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", SlashCommandPath))

	// Validate and parse the inbound request.
	body := checkRequest(w, r, l, api.ContentTypeForm)
	if body == nil {
		return
	}

	kv, err := url.ParseQuery(string(body))
	if err != nil {
		l.Error("Failed to parse slash command's URL-encoded form",
			zap.Error(err),
			zap.ByteString("body", body),
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

	// Transform the received Slack event into an autokitteh event.
	data, err := transformCommand(l, w, cmd)
	if err != nil {
		return
	}
	akEvent := &sdktypes.EventPB{
		IntegrationId:   h.integrationID.String(),
		OriginalEventId: cmd.TriggerID,
		EventType:       "slash_command",
		Data:            data,
	}

	// Retrieve all the relevant connections for this event.
	connTokens, err := h.listTokens(cmd.APIAppID, cmd.EnterpriseID, cmd.TeamID)
	if err != nil {
		l.Error("Failed to retrieve connection tokens",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(l, connTokens, akEvent)

	// https://api.slack.com/interactivity/slash-commands#responding_to_commands
	// https://api.slack.com/interactivity/slash-commands#responding_response_url
	// https://api.slack.com/interactivity/slash-commands#enabling-interactivity-with-slash-commands__best-practices
	if len(cmd.Text) == 0 {
		return
	}
	w.Header().Add(api.HeaderContentType, api.ContentTypeJSONCharsetUTF8)
	fmt.Fprintf(w, "{\"response_type\": \"ephemeral\", \"text\": \"Your command: `%s`\"}", cmd.Text)
}

// transformCommand transforms a received Slack event into an autokitteh event.
func transformCommand(l *zap.Logger, w http.ResponseWriter, cmd SlashCommand) (map[string]*valuesv1.Value, error) {
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(cmd)
	if err != nil {
		l.Error("Failed to wrap Slack event",
			zap.Error(err),
			zap.Any("cmd", cmd),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Slack event",
			zap.Error(err),
			zap.Any("cmd", cmd),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	return kittehs.TransformMapValues(data, sdktypes.ToProto), nil
}
