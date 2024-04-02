package websockets

import (
	"context"
	"encoding/json"

	"github.com/slack-go/slack/socketmode"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/events"
	"go.autokitteh.dev/autokitteh/integrations/slack/webhooks"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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

	// Transform the received Slack event into an autokitteh event.
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(slackEvent)
	if err != nil {
		h.logger.Error("Failed to wrap Slack event",
			zap.Any("cmd", slackEvent),
			zap.Error(err),
		)
		return
	}

	m, err := wrapped.ToStringValuesMap()
	if err != nil {
		h.logger.Error("Failed to convert wrapped Slack event",
			zap.Any("event", slackEvent),
			zap.Error(err),
		)
		return
	}

	pb := kittehs.TransformMapValues(m, sdktypes.ToProto)
	akEvent := &sdktypes.EventPB{
		IntegrationId: h.integrationID.String(),
		EventType:     cb.Event.Type,
		Data:          pb,
	}

	// Retrieve all the relevant connections for this event.
	connTokens, err := h.secrets.List(context.Background(), h.scope, "websockets")
	if err != nil {
		h.logger.Error("Failed to retrieve connection tokens", zap.Error(err))
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(connTokens, akEvent)
}
