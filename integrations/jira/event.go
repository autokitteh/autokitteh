package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

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

type event struct {
	WebhookEvent      string `json:"webhookEvent"`
	MatchedWebhookIDs []int  `json:"matchedWebhookIds"`
	Timestamp         int64  `json:"timestamp"`
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

	// Verify the JWT in the event's "Authorization" header.
	token := r.Header.Get(headerAuthorization)
	if !verifyJWT(l, strings.TrimPrefix(token, "Bearer ")) {
		// This is probably an attack, so no user-friendliness.
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeJSON) {
		// This is probably an attack, so no user-friendliness.
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse some of the metadata in the Jira event's JSON content.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// This is probably an attack, so no user-friendliness.
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var e event
	if err := json.Unmarshal(body, &e); err != nil {
		// This is probably an attack, so no user-friendliness.
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Construct an AutoKitteh event from the Jira event.
	event, err := constructEvent(l, body, e)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	// Iterate through all the relevant connections for this event.
	ctx := r.Context()
	key := sdktypes.NewSymbol("webhook_id")
	for _, id := range e.MatchedWebhookIDs {
		value := fmt.Sprintf("%d", id)
		cids, err := h.vars.FindConnectionIDs(ctx, integrationID, key, value)
		if err != nil {
			l.Error("Failed to find connection IDs", zap.Error(err))
			continue
		}

		// Dispatch the event to all of them, for asynchronous handling.
		h.dispatchAsyncEventsToConnections(ctx, l, cids, event)
	}

	// Returning immediately without an error = acknowledgement of receipt.
}

// https://developer.atlassian.com/cloud/jira/platform/understanding-jwt-for-connect-apps/
func verifyJWT(l *zap.Logger, authz string) bool {
	token, err := jwt.Parse(authz, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			l.Warn("Unexpected signing method", zap.Any("alg", token.Header["alg"]))
		}
		// TODO(ENG-965): From new-connection form instead of env vars.
		return []byte(os.Getenv("JIRA_CLIENT_SECRET")), nil
	})
	if err != nil {
		l.Warn("Failed to parse JWT", zap.Error(err))
		return false
	}

	return token.Valid
}

func constructEvent(l *zap.Logger, body []byte, e event) (*sdktypes.EventPB, error) {
	m := map[string]string{"json": string(body)}
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(m)
	if err != nil {
		l.Error("Failed to wrap Jira event", zap.Any("event", body), zap.Error(err))
		return nil, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Jira event", zap.Any("event", body), zap.Error(err))
		return nil, err
	}

	return &sdktypes.EventPB{
		EventType: strings.TrimPrefix(e.WebhookEvent, "jira:"),
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	}, nil
}

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, l *zap.Logger, cids []sdktypes.ConnectionID, event *sdktypes.EventPB) {
	for _, cid := range cids {
		event.ConnectionId = cid.String()
		e, err := sdktypes.EventFromProto(event)
		if err != nil {
			l.Error("Failed to convert protocol buffer to SDK event", zap.Error(err))
			return
		}

		eventID, err := h.dispatcher.Dispatch(ctx, e, nil)
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
