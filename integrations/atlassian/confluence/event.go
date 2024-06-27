package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType   = "Content-Type"
	headerAuthorization = "Authorization"
	contentTypeJSON     = "application/json"
)

// Default HTTP client with a timeout for short-lived HTTP requests.
var httpClient = http.Client{Timeout: 3 * time.Second}

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger     *zap.Logger
	oauth      sdkservices.OAuth
	vars       sdkservices.Vars
	dispatcher sdkservices.Dispatcher
}

func NewHTTPHandler(l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars, d sdkservices.Dispatcher) handler {
	return handler{logger: l, oauth: o, vars: v, dispatcher: d}
}

// handleEvent receives from Atlassian asynchronous events,
// and dispatches them to zero or more AutoKitteh connections.
// Note 1: By default, AutoKitteh creates webhooks automatically,
// subscribing to all events - see "webhooks.go" for more details.
// Note 2: Unlike Jira webhooks, auto-created Confluence webhooks
// do not expire after 30 days, so no need to extend their deadline.
// Note 3: The requests are sent by a service, so no need to respond
// with user-friendly error web pages.
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// TODO(ENG-1081): Verify the HMAC signature in "X-Hub-Signature"
	// (Confluence doesn't recognize secrets when creating webhooks).

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeJSON) {
		l.Warn("Incoming Atlassian event with bad header", zap.String(headerContentType, contentType))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse some of the metadata in the Atlassian event's JSON content.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Warn("Failed to read content of incoming Atlassian event", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var atlassianEvent map[string]any
	if err := json.Unmarshal(body, &atlassianEvent); err != nil {
		l.Warn("Failed to unmarshal JSON in incoming Atlassian event", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Determine the event type based on the webhook callback URL, and the content
	// of the event (unlike Jira, Confluence events don't have an event type field).
	eventType := r.PathValue("category")
	if eventType == "" {
		l.Warn("Unexpected Confluence event callback without category in URL")
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}

	// TODO(ENG-1081): Parse event content, add as a prefix to the category.

	// Construct an AutoKitteh event from the Atlassian event.
	akEvent, err := constructEvent(l, atlassianEvent, eventType)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// TODO(ENG-1081): Iterate through all the relevant connections for this event.
	// is, ok := atlassianEvent["matchedWebhookIds"].([]any)
	// if !ok {
	// 	l.Warn("Invalid webhook IDs in Atlassian event", zap.Any("ids", atlassianEvent["matchedWebhookIds"]))
	// 	http.Error(w, "Bad Request", http.StatusBadRequest)
	// 	return
	// }

	ids := []int{}
	// ids := kittehs.Transform(is, func(v any) int {
	// 	f, ok := v.(float64)
	// 	if !ok {
	// 		l.Warn("Invalid webhook ID in Atlassian event", zap.Any("id", v))
	// 		return 0
	// 	}
	// 	return int(f)
	// })

	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	for _, id := range ids {
		value := fmt.Sprintf("%d", id)
		cids, err := h.vars.FindConnectionIDs(ctx, integrationID, webhookID, value)
		if err != nil {
			l.Error("Failed to find connection IDs", zap.Error(err))
			continue
		}

		// Dispatch the event to all of them, for asynchronous handling.
		h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)
	}

	// Returning immediately without an error = acknowledgement of receipt.
}

func constructEvent(l *zap.Logger, atlassianEvent map[string]any, eventType string) (sdktypes.Event, error) {
	l = l.With(zap.Any("event", atlassianEvent))

	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(atlassianEvent)
	if err != nil {
		l.Error("Failed to wrap Atlassian event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Atlassian event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.String("eventType", eventType),
			zap.Any("data", data),
			zap.Error(err),
		)
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	for _, cid := range cids {
		eid, err := h.dispatcher.Dispatch(ctx, e.WithConnectionID(cid), nil)
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
