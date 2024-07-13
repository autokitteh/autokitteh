package forms

import (
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger     *zap.Logger
	vars       sdkservices.Vars
	dispatcher sdkservices.Dispatcher
}

func NewWebhookHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.Dispatcher) http.Handler {
	return handler{logger: l, vars: v, dispatcher: d}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	eventType := r.Header.Get("Eventtype")
	formID := r.Header.Get("Formid")
	publishTime := r.Header.Get("X-Goog-Pubsub-Publish-Time")

	l := h.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("eventType", eventType),
		zap.String("formID", formID),
		zap.String("watchID", r.Header.Get("Watchid")),
		zap.String("messageID", r.Header.Get("X-Goog-Pubsub-Message-Id")),
		zap.String("publishTime", publishTime),
		zap.String("subscriptionName", r.Header.Get("X-Goog-Pubsub-Subscription-Name")),
	)
	l.Info("Received Google Forms notification")

	switch WatchEventType(eventType) {
	case WatchSchemaChanges:
		// TODO(ENG-1103)

	case WatchNewResponses:
		// TODO(ENG-1103)

	default:
		l.Error("Unknown Google Forms event type")
		return
	}

	// TODO(ENG-1103): get form/response...
}
