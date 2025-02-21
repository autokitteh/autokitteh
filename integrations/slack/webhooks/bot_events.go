package webhooks

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
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
	"app_uninstalled": events.AppUninstalledTokensRevokedHandler,

	"channel_archive":   events.ChannelGroupMemberHandler,
	"channel_created":   events.ChannelCreatedRenameHandler,
	"channel_left":      events.ChannelGroupMemberHandler,
	"channel_rename":    events.ChannelCreatedRenameHandler,
	"channel_unarchive": events.ChannelGroupMemberHandler,

	"group_archive":   events.ChannelGroupMemberHandler,
	"group_close":     events.ChannelGroupMemberHandler,
	"group_deleted":   events.ChannelGroupMemberHandler,
	"group_left":      events.ChannelGroupMemberHandler,
	"group_open":      events.ChannelGroupMemberHandler,
	"group_rename":    events.ChannelCreatedRenameHandler,
	"group_unarchive": events.ChannelGroupMemberHandler,

	"member_joined_channel": events.ChannelGroupMemberHandler,
	"member_left_channel":   events.ChannelGroupMemberHandler,

	"message": events.MessageHandler,

	"reaction_added":   events.ReactionHandler,
	"reaction_removed": events.ReactionHandler,

	"tokens_revoked": events.AppUninstalledTokensRevokedHandler,

	"url_verification": events.URLVerificationHandler,
}

// HandleBotEvent routes all asynchronous bot event notifications that our
// app subscribed to, to specific event handlers based on the event type.
// See https://api.slack.com/apis/connections/events-api#responding.
// Compare with the [websockets.handleBotEvent] implementation.
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
		common.HTTPError(w, http.StatusBadRequest)
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
		common.HTTPError(w, http.StatusNotImplemented)
		return
	}
	slackEvent := f(l, w, body, cb)
	if slackEvent == nil {
		return
	}

	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(l, slackEvent, cb.Event.Type)
	if err != nil {
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	enterpriseID := "" // TODO: Support enterprise IDs.
	cids, err := h.listConnectionIDs(ctx, cb.APIAppID, enterpriseID, cb.TeamID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)

	// Returning immediately without an error = acknowledgement of receipt.
}
