package google

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handlNotif receives and dispatches asynchronous Google Forms
// notifications from a push subscription to a GCP Cloud Pub/Sub topic.
func (h handler) handleFormsNotif(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("eventType", r.Header.Get("Eventtype")),
		zap.String("formID", r.Header.Get("Formid")),
		zap.String("watchID", r.Header.Get("Watchid")),
		zap.String("messageID", r.Header.Get("X-Goog-Pubsub-Message-Id")),
		zap.String("publishTime", r.Header.Get("X-Goog-Pubsub-Publish-Time")),
		zap.String("subscriptionName", r.Header.Get("X-Goog-Pubsub-Subscription-Name")),
	)
	l.Info("Received Google Forms notification")

	// Parse event details from the request headers.
	eventType := forms.WatchEventType(r.Header.Get("Eventtype"))
	watchID := r.Header.Get("Watchid")

	name := vars.FormResponsesWatchID
	if eventType == forms.WatchSchemaChanges {
		name = vars.FormSchemaWatchID
	}

	// Find all the connection IDs associated with the watch ID.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	cids, err := h.vars.FindConnectionIDs(ctx, integrationID, name, watchID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Construct the event and dispatch it to all the connections.
	formsEvent := map[string]any{
		"event_type":   r.Header.Get("Eventtype"),
		"form_id":      r.Header.Get("Formid"),
		"publish_time": r.Header.Get("X-Goog-Pubsub-Publish-Time"),
	}

	akEvent, err := forms.ConstructEvent(ctx, h.vars, formsEvent, cids)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.dispatchAsyncEventsToConnections(ctx, cids, akEvent); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Returning immediately without an error = acknowledgement of receipt.
}

type gmailNotifBody struct {
	EmailAddress string `json:"emailAddress"`
	HistoryID    int    `json:"historyId"`
}

// handlNotif receives and dispatches asynchronous Gmail
// notifications from a push subscription to a GCP Cloud Pub/Sub topic.
func (h handler) handleGmailNotif(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("messageID", r.Header.Get("X-Goog-Pubsub-Message-Id")),
		zap.String("publishTime", r.Header.Get("X-Goog-Pubsub-Publish-Time")),
		zap.String("subscriptionName", r.Header.Get("X-Goog-Pubsub-Subscription-Name")),
	)
	l.Info("Received Gmail notification")

	// Parse event details from the JSON body.
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Warn("Failed to read request body", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	notif := gmailNotifBody{}
	if err := json.Unmarshal(body, &notif); err != nil {
		l.Warn("Failed to unmarshal request body", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	l = l.With(
		zap.String("emailAddress", notif.EmailAddress),
		zap.Int("historyID", notif.HistoryID),
	)

	// Find all the connection IDs associated with the email address.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	cids, err := h.vars.FindConnectionIDs(ctx, integrationID, vars.UserEmail, notif.EmailAddress)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Construct the event and dispatch it to all the connections.
	gmailEvent := map[string]any{
		"publish_time":  r.Header.Get("X-Goog-Pubsub-Publish-Time"),
		"email_address": notif.EmailAddress,
		"history_id":    notif.HistoryID,
	}

	akEvent, err := gmail.ConstructEvent(ctx, h.vars, gmailEvent, cids)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if err := h.dispatchAsyncEventsToConnections(ctx, cids, akEvent); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Returning immediately without an error = acknowledgement of receipt.
}

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) error {
	l := extrazap.ExtractLoggerFromContext(ctx)

	for _, cid := range cids {
		eid, err := h.dispatcher.Dispatch(ctx, e.WithConnectionID(cid), nil)
		l := l.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return err
		}
		l.Debug("Event dispatched")
	}

	return nil
}
