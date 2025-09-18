package webhooks

import (
	"net/http"

	"github.com/iancoleman/strcase"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
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

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse inbound HTTP request", zap.Error(err))
		// Attack or network loss, so no need for user-friendliness.
		common.HTTPError(w, http.StatusBadRequest)
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

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: "message",
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.Any("data", data),
			zap.Error(err),
		)
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	cids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, sdktypes.NewSymbol("account_sid"), aid)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)
}
