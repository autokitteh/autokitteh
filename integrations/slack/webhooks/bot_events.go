package webhooks

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/events"
)

const (
	BotEventPath = "/slack/event"
)

type botEventHandler = func(*zap.Logger, http.ResponseWriter, []byte, *events.Callback) any

var botEventHandlers = map[string]botEventHandler{
	"app_mention": events.AppMentionHandler,
	// TODO: app_rate_limit
	// TODO: app_uninstalled

	"channel_archive": events.ChannelGroupHandler,
	"channel_created": events.ChannelCreatedHandler,
	// TODO: channel_deleted
	// TODO: channel_history_changed
	// TODO: channel_id_changed
	// TODO: channel_left
	// TODO: channel_rename
	"channel_unarchive": events.ChannelGroupHandler,

	"group_archive": events.ChannelGroupHandler,
	// TODO: group_deleted
	// TODO: group_history_changed
	// TODO: group_left
	"group_open": events.ChannelGroupHandler,
	// TODO: group_rename
	"group_unarchive": events.ChannelGroupHandler,

	// TODO: im_history_changed

	"member_joined_channel": events.ChannelGroupHandler,

	"message": events.MessageHandler,

	"reaction_added":   events.ReactionHandler,
	"reaction_removed": events.ReactionHandler,

	// TODO: tokens_revoked

	"url_verification": events.URLVerificationHandler,
}

// HandleBotEvent routes all asynchronous bot event notifications that our Slack
// app subscribed to, to specific event handlers based on the event type.
// See https://api.slack.com/apis/connections/events-api#responding.
func (h handler) HandleBotEvent(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", BotEventPath))

	// Validate and parse the inbound request.
	body := checkRequest(w, r, l, api.ContentTypeJSON)
	if body == nil {
		return
	}

	cb := &events.Callback{}
	if err := json.Unmarshal(body, cb); err != nil {
		l.Error("Failed to parse bot event's JSON payload",
			zap.Error(err),
			zap.ByteString("json", body),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse the received event's inner details based on its type.
	t := cb.Type
	if t == "event_callback" {
		t = cb.Event.Type
	}
	l = l.With(zap.String("event", t))
	f, ok := botEventHandlers[t]
	if !ok {
		l.Error("Received unsupported bot event",
			zap.ByteString("body", body),
			zap.Any("callback", cb),
		)
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
		return
	}
	slackEvent := f(l, w, body, cb)
	if slackEvent == nil {
		return
	}

	// Transform the received Slack event into an autokitteh event.
	data, err := transformEvent(l, w, slackEvent)
	if err != nil {
		return
	}
	akEvent := &sdktypes.EventPB{
		IntegrationId:   h.integrationID.String(),
		OriginalEventId: cb.EventID,
		EventType:       cb.Event.Type,
		Data:            data,
	}

	// Retrieve all the relevant connections for this event.
	enterpriseID := "" // TODO: Support enterprise IDs.
	connTokens, err := h.listTokens(cb.APIAppID, enterpriseID, cb.TeamID)
	if err != nil {
		l.Error("Failed to retrieve connection tokens",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(l, connTokens, akEvent)

	// Returning immediately without an error = acknowledgement of receipt.
}

// transformEvent transforms a received Slack event into an autokitteh event.
func transformEvent(l *zap.Logger, w http.ResponseWriter, event any) (map[string]*valuesv1.Value, error) {
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(event)
	if err != nil {
		l.Error("Failed to wrap Slack event",
			zap.Error(err),
			zap.Any("innerEvent", event),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Slack event",
			zap.Error(err),
			zap.Any("innerEvent", event),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	return kittehs.TransformMapValues(data, sdktypes.ToProto), nil
}
