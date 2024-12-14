package google

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/google/calendar"
	"go.autokitteh.dev/autokitteh/integrations/google/drive"
	"go.autokitteh.dev/autokitteh/integrations/google/forms"
	"go.autokitteh.dev/autokitteh/integrations/google/gmail"
	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleCalNotification receives and dispatches asynchronous Google Calendar notifications.
func (h handler) handleCalNotification(w http.ResponseWriter, r *http.Request) {
	// Parse event details from the request headers.
	channelID := r.Header.Get("X-Goog-Channel-Id")
	resState := r.Header.Get("X-Goog-Resource-State")

	l := h.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("channelID", channelID),
		zap.String("channelToken", r.Header.Get("X-Goog-Channel-Token")),
		zap.String("resourceID", r.Header.Get("X-Goog-Resource-Id")),
		zap.String("resourceState", resState),
		zap.String("messageNumber", r.Header.Get("X-Goog-Message-Number")),
	)
	if resState == "sync" {
		l.Info("Ignoring Google Calendar watch creation notification")
		return
	}
	l.Info("Received Google Calendar notification")

	// Find all the connection IDs associated with the watch ID.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	name := vars.CalendarEventsWatchID
	cids, err := h.vars.FindConnectionIDs(ctx, calendar.IntegrationID, name, channelID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Construct the event and dispatch it to all the connections.
	akEvent, err := calendar.ConstructEvent(ctx, h.vars, cids)
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

// handleDriveNotification receives and dispatches asynchronous Google Drive notifications.
func (h handler) handleDriveNotification(w http.ResponseWriter, r *http.Request) {
	// Parse event details from the request headers
	channelID := r.Header.Get("X-Goog-Channel-Id")
	resState := r.Header.Get("X-Goog-Resource-State")

	l := h.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("channelID", channelID),
		zap.String("channelToken", r.Header.Get("X-Goog-Channel-Token")),
		zap.String("resourceID", r.Header.Get("X-Goog-Resource-Id")),
		zap.String("resourceState", resState),
		zap.String("messageNumber", r.Header.Get("X-Goog-Message-Number")),
	)

	if resState == "sync" {
		l.Info("Ignoring Google Drive watch creation notification")
		return
	}
	l.Info("Received Google Drive notification")

	// Find all the connection IDs associated with the watch ID.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	name := vars.DriveEventsWatchID
	cids, err := h.vars.FindConnectionIDs(ctx, drive.IntegrationID, name, channelID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	akEvent, err := drive.ConstructEvent(ctx, h.vars, cids)
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

// handleFormsNotification receives and dispatches asynchronous Google Forms
// notifications from a push subscription to a GCP Cloud Pub/Sub topic.
func (h handler) handleFormsNotification(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))
	if !checkRequest(w, r, l) {
		return
	}

	// Parse event details from the request headers.
	eventType := forms.WatchEventType(r.Header.Get("Eventtype"))
	watchID := r.Header.Get("Watchid")

	name := vars.FormResponsesWatchID
	if eventType == forms.WatchSchemaChanges {
		name = vars.FormSchemaWatchID
	}

	l = l.With(
		zap.String("eventType", r.Header.Get("Eventtype")),
		zap.String("formID", r.Header.Get("Formid")),
		zap.String("watchID", watchID),
		zap.String("messageID", r.Header.Get("X-Goog-Pubsub-Message-Id")),
		zap.String("publishTime", r.Header.Get("X-Goog-Pubsub-Publish-Time")),
		zap.String("subscriptionName", r.Header.Get("X-Goog-Pubsub-Subscription-Name")),
	)
	l.Info("received Google Forms notification")

	// Find all the connection IDs associated with the watch ID.
	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	cids, err := h.vars.FindConnectionIDs(ctx, forms.IntegrationID, name, watchID)
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

// handleGmailNotification receives and dispatches asynchronous Gmail
// notifications from a push subscription to a GCP Cloud Pub/Sub topic.
func (h handler) handleGmailNotification(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))
	if !checkRequest(w, r, l) {
		return
	}

	l = l.With(
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
	cids, err := h.vars.FindConnectionIDs(ctx, gmail.IntegrationID, vars.UserEmail, notif.EmailAddress)
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
		eid, err := h.dispatcher.Dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
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

// checkRequest checks the authenticity of the given HTTP POST request.
// If the check fails, it also sets an HTTP error to the impostor client.
func checkRequest(w http.ResponseWriter, r *http.Request, l *zap.Logger) bool {
	// Read the bearer token from the Authorization request header.
	auth := r.Header.Get("Authorization")
	if auth == "" {
		l.Warn("missing authorization header in Google push notification")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	if !strings.HasPrefix(auth, "Bearer ") {
		l.Warn("invalid authorization header in Google push notification")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}

	auth = strings.TrimPrefix(auth, "Bearer ")

	// Download and cache Google's OAuth public keys.
	rsaPublicKeys, err := fetchGoogleCerts()
	if err != nil {
		l.Error("failed to fetch Google OAuth certs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return false
	}

	// Parse the JWT and verify its signature.
	token, err := jwt.Parse(auth, func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing or invalid kid in token header")
		}

		key, exists := rsaPublicKeys[kid]
		if !exists {
			return nil, fmt.Errorf("Google public key not found for kid: %s", kid)
		}
		return key, nil
	})
	if err != nil {
		l.Error("failed to parse JWT in Google push notification", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return false
	}

	return token.Valid
}

const (
	googleCertsURL = "https://www.googleapis.com/oauth2/v3/certs"
	cacheTimeout   = 24 * time.Hour
)

var (
	cachedPublicKeys map[string]*rsa.PublicKey
	cacheDeadline    = time.Now()
)

// fetchGoogleCerts downloads Google's OAuth public keys and caches them.
func fetchGoogleCerts() (map[string]*rsa.PublicKey, error) {
	// Return the cached results if they're still fresh.
	if time.Now().Before(cacheDeadline) {
		return cachedPublicKeys, nil
	}

	// Download JSON.
	resp, err := http.Get(googleCertsURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching Google OAuth certs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching Google OAuth certs resulted in status %d", resp.StatusCode)
	}

	// Read JSON.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading Google OAuth certs: %w", err)
	}

	var certs struct {
		Keys []struct {
			KID string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	// Parse JSON.
	if err := json.Unmarshal(body, &certs); err != nil {
		return nil, fmt.Errorf("error unmarshaling Google OAuth certs: %w", err)
	}

	cachedPublicKeys = make(map[string]*rsa.PublicKey)
	for _, key := range certs.Keys {
		n, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			return nil, fmt.Errorf("error decoding modulus %q: %w", key.N, err)
		}

		e, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			return nil, fmt.Errorf("error decoding exponent %q: %w", key.N, err)
		}

		pk := &rsa.PublicKey{
			N: new(big.Int).SetBytes(n),
			E: int(new(big.Int).SetBytes(e).Uint64()),
		}
		cachedPublicKeys[key.KID] = pk
	}

	cacheDeadline = time.Now().Add(cacheTimeout)
	return cachedPublicKeys, nil
}
