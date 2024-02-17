package webhooks

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
)

// handler is an autokitteh webhook which implements [http.Handler] to
// receive, dispatch, and acknowledge asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	secrets       sdkservices.Secrets
	dispatcher    sdkservices.Dispatcher
	scope         string
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, sec sdkservices.Secrets, d sdkservices.Dispatcher, scope string, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		secrets:       sec,
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

// listTokens calls the List method in SecretsService.
func (h handler) listTokens(accountSID string) ([]string, error) {
	key := fmt.Sprintf("accounts/%s", accountSID)
	tokens, err := h.secrets.List(context.Background(), h.scope, key)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (h handler) dispatchAsyncEventsToConnections(l *zap.Logger, tokens []string, event *eventsv1.Event) {
	ctx := extrazap.AttachLoggerToContext(l, context.Background())
	for _, connToken := range tokens {
		event.IntegrationToken = connToken
		event := kittehs.Must1(sdktypes.EventFromProto(event))
		eventID, err := h.dispatcher.Dispatch(ctx, event, nil)
		if err != nil {
			l.Error("Dispatch failed",
				zap.String("connectionToken", connToken),
				zap.Error(err),
			)
			return
		}
		l.Debug("Dispatched",
			zap.String("connectionToken", connToken),
			zap.String("eventID", eventID.String()),
		)
	}
}
