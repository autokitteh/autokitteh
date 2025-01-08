package webhooks

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/events"
)

const (
	BotEventPath = "/slack/event"
)

type BotEventHandler = func(*zap.Logger, http.ResponseWriter, []byte, *events.Callback) any

var BotEventHandlers = map[string]BotEventHandler{
	"app_home_opened": events.AppHomeOpenedHandler,
	"app_mention":     events.AppMentionHandler,
	// TODO: app_rate_limit
	// TODO: app_uninstalled

	"channel_archive": events.ChannelGroupMemberHandler,
	"channel_created": events.ChannelCreatedRenameHandler,
	// TODO: channel_history_changed
	// TODO: channel_id_changed
	"channel_rename":    events.ChannelCreatedRenameHandler,
	"channel_unarchive": events.ChannelGroupMemberHandler,

	"group_archive": events.ChannelGroupMemberHandler,
	// TODO: group_history_changed
	"group_open":      events.ChannelGroupMemberHandler,
	"group_rename":    events.ChannelCreatedRenameHandler,
	"group_unarchive": events.ChannelGroupMemberHandler,

	// TODO: im_history_changed

	"member_joined_channel": events.ChannelGroupMemberHandler,
	"member_left_channel":   events.ChannelGroupMemberHandler,

	"message": events.MessageHandler,

	"reaction_added":   events.ReactionHandler,
	"reaction_removed": events.ReactionHandler,

	// TODO: tokens_revoked

	"url_verification": events.URLVerificationHandler,
}

// HandleBotEvent routes all asynchronous bot event notifications that our Slack
// app subscribed to, to specific event handlers based on the event type.
// See https://api.slack.com/apis/connections/events-api#responding.
// Compare this function with the [websockets.HandleBotEvent] implementation.
func (h handler) HandleBotEvent(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", BotEventPath))

	// Validate and parse the inbound request.
	body := h.checkRequest(w, r, l, api.ContentTypeJSON)
	if body == nil {
		return
	}

	cb := &events.Callback{}
	if err := json.Unmarshal(body, cb); err != nil {
		l.Error("Failed to parse bot event's JSON payload",
			zap.ByteString("json", body),
			zap.Error(err),
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
	f, ok := BotEventHandlers[t]
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

	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(l, slackEvent, cb.Event.Type)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	enterpriseID := "" // TODO: Support enterprise IDs.
	cids, err := h.listConnectionIDs(ctx, cb.APIAppID, enterpriseID, cb.TeamID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)

	// Returning immediately without an error = acknowledgement of receipt.
}
