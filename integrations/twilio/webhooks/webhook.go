package webhooks

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler] to
// receive, dispatch, and acknowledge asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatcher    sdkservices.Dispatcher
	scope         string
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, vars sdkservices.Vars, d sdkservices.Dispatcher, scope string, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		vars:          vars,
		dispatcher:    d,
		scope:         scope,
		integrationID: id,
	}
}

// TODO:

// https://www.twilio.com/docs/usage/webhooks/webhooks-overview
// https://www.twilio.com/docs/usage/webhooks/getting-started-twilio-webhooks
// https://www.twilio.com/docs/messaging/guides/webhook-request

// https://www.twilio.com/docs/usage/webhooks/sms-webhooks
// https://www.twilio.com/docs/messaging/tutorials/how-to-confirm-delivery/python
// https://www.twilio.com/docs/messaging/tutorials/how-to-receive-and-reply/python
// https://www.twilio.com/docs/messaging/twiml

// https://www.twilio.com/docs/usage/troubleshooting/debugging-event-webhooks
// https://www.twilio.com/docs/usage/api/usage-trigger
// https://www.twilio.com/docs/usage/troubleshooting/alarms

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, l *zap.Logger, cids []sdktypes.ConnectionID, event *sdktypes.EventPB) {
	for _, cid := range cids {
		event.ConnectionId = cid.String()

		event, err := sdktypes.EventFromProto(event)
		if err != nil {
			l.Error("Failed to convert protocol buffer to SDK event", zap.Error(err))
			return
		}

		eventID, err := h.dispatcher.Dispatch(ctx, event, nil)
		if err != nil {
			l.Error("Event dispatch failed",
				zap.String("eventID", eventID.String()),
				zap.String("connectionID", cid.String()),
				zap.Error(err),
			)
			return
		}
		l.Debug("Event dispatched",
			zap.String("eventID", eventID.String()),
			zap.String("connectionID", cid.String()),
		)

	}
}
