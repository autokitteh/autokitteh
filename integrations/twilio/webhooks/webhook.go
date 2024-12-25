package webhooks

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
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
	integration   sdktypes.Integration
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, vars sdkservices.Vars, d sdkservices.Dispatcher, scope string, i sdktypes.Integration) handler {
	return handler{
		logger:        l,
		vars:          vars,
		dispatcher:    d,
		scope:         scope,
		integration:   i,
		integrationID: i.ID(),
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

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	for _, cid := range cids {
		eid, err := h.dispatcher.Dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		l := l.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("Event dispatched")
	}
}
