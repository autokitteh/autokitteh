package webhooks

import (
	"fmt"
	"net/http"

	"github.com/iancoleman/strcase"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// MessagePath is the URL path for our webhook to handle message events.
	MessagePath = "/twilio/message"
)

// HandleMessage dispatches to autokitteh an asynchronous event notification.
func (h handler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", MessagePath))

	// TODO: Validate and parse the inbound request.
	// https://www.twilio.com/docs/usage/webhooks/getting-started-twilio-webhooks#validate-that-webhook-requests-are-coming-from-twilio
	// https://www.twilio.com/docs/usage/webhooks/webhooks-security
	// https://www.twilio.com/docs/usage/security#validating-requests
	// https://www.twilio.com/docs/usage/tutorials/how-to-secure-your-gin-project-by-validating-incoming-twilio-requests

	if err := r.ParseForm(); err != nil {
		l.Warn("parse form error", zap.Error(err))
		return
	}

	aid := r.Form.Get("AccountSid")
	mid := r.Form.Get("MessageSid")
	l = l.With(zap.String("accountSID", aid), zap.String("messageSID", mid))

	// Transform the received Twilio event into an autokitteh event.
	data := make(map[string]sdktypes.Value, len(r.Form))
	for k, vs := range r.Form {
		if len(vs) > 0 {
			data[strcase.ToSnake(k)] = sdktypes.NewStringValue(vs[0])
		}
	}
	akEvent := &sdktypes.EventPB{
		IntegrationId:   h.integrationID.String(),
		OriginalEventId: fmt.Sprintf("%s/%s", aid, mid),
		EventType:       "message",
		Data:            kittehs.TransformMapValues(data, sdktypes.ToProto),
	}

	// Retrieve all the relevant connections for this event.
	connTokens, err := h.listTokens(aid)
	if err != nil {
		l.Error("Failed to retrieve connection tokens",
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(l, connTokens, akEvent)
}
