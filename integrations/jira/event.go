package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

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

type event struct {
	MatchedWebhookIDs []int `json:"matchedWebhookIds"`
}

// handleEvent receives from Jira asynchronous events,
// and dispatches them to one or more AutoKitteh connections.
// Note 1: By default, AutoKitteh creates webhooks automatically,
// subscribing to all events - see "webhooks.go" for more details.
// TODO(ENG-965):
// Note 2: Dynamic (i.e. auto-created) webhooks expire after 30 days.
// This functions extends this deadline at the 20-day mark.
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check the "Content-Type" header.
	header := r.Header.Get(headerContentType)
	if !strings.HasPrefix(header, contentTypeJSON) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse the event's JSON content, specifically the webhook ID(s).
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var e event
	if err := json.Unmarshal(body, &e); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	key := sdktypes.NewSymbol("webhook_id")
	for _, id := range e.MatchedWebhookIDs {
		l.Warn("Webhook ID", zap.Int("id", id)) // TODO: REMOVE THIS LINE!

		// TODO: Check the "Authorization" header.
		header = r.Header.Get(headerAuthorization)
		l.Warn("Authorization header", zap.String("header", header)) // TODO: REMOVE THIS LINE!

		// Retrieve all the relevant connections for this event.
		value := fmt.Sprintf("%d", id)
		cs, err := h.vars.FindConnectionIDs(ctx, integrationID, key, value)
		if err != nil {
			l.Warn("Failed to find connection IDs",
				zap.String("jiraWebHookID", value),
				zap.Error(err),
			)
			continue
		}

		// TODO: Dispatch the event to all of them, for asynchronous handling.
		for _, c := range cs {
			l.Warn("Connection ID", zap.String("cid", c.String())) // TODO: REMOVE THIS LINE!
		}
	}

	// Returning immediately without an error = acknowledgement of receipt.
}
