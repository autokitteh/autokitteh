package websockets

import (
	"context"
	"encoding/json"

	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/events"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
)

// HandleBotEvent routes all asynchronous bot event notifications that our Slack
// app subscribed to, to specific event handlers based on the event type.
// See https://api.slack.com/apis/connections/events-api#responding.
// Compare this function with the [webhooks.HandleBotEvent] implementation.
func (h handler) handleBotEvent(e *socketmode.Event, c *socketmode.Client) {
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
	cb := &events.Callback{}
	if err := json.Unmarshal(body, cb); err != nil {
		h.logger.Error("Failed to parse bot event's JSON payload",
			zap.ByteString("json", body),
			zap.Error(err),
		)
		return
	}

	// Parse the received event's inner details based on its type.
	t := cb.Type
	if t == "event_callback" {
		t = cb.Event.Type
	}
	l := h.logger.With(zap.String("event", t))
	f, ok := webhooks.BotEventHandlers[t]
	if !ok {
		l.Error("Received unsupported bot event",
			zap.ByteString("body", body),
			zap.Any("callback", cb),
		)
		return
	}
	slackEvent := f(l, nil, body, cb)
	if slackEvent == nil {
		return
	}

	// Transform the received Slack event into an AutoKitteh event.
	akEvent, err := transformEvent(h.logger, slackEvent, cb.Event.Type)
	if err != nil {
		return
	}

	// Retrieve all the relevant connections for this event.
	cids, err := h.vars.FindConnectionIDs(context.Background(), h.integrationID, vars.AppTokenName, "")
	if err != nil {
		h.logger.Error("Failed to find connection IDs", zap.Error(err))
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(cids, akEvent)
}
