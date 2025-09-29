package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// headerSecretToken is the HTTP header that contains the secret token for webhook verification
	headerSecretToken = "X-Telegram-Bot-Api-Secret-Token"
)

// handleEvent processes incoming Telegram webhook events.
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check the request's headers and parse its body.
	telegramEvent := h.checkRequest(w, r)
	if telegramEvent == nil {
		return
	}

	// Transform the Telegram event into an AutoKitteh event.
	eventType := getTelegramEventType(telegramEvent)
	akEvent, err := common.TransformEvent(h.logger, telegramEvent, eventType)
	if err != nil {
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	// Retrieve all the relevant connections for this event.
	cids, err := h.findAuthenticatedConnections(ctx, r, h.logger)
	if err != nil {
		common.HTTPError(w, http.StatusUnauthorized)
		return
	}

	// Dispatch the event to all of them, for potential asynchronous handling.
	common.DispatchEvent(ctx, h.logger, h.dispatch, akEvent, cids)
}

func (h handler) checkRequest(w http.ResponseWriter, r *http.Request) map[string]any {
	l := h.logger.With(
		zap.String("url_path", r.URL.Path),
		zap.String("event_type", r.Header.Get("Telegram-Event")),
	)

	// Check the request's HTTP headers.
	if common.PostWithoutJSONContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("incoming event: unexpected content type", zap.String("content_type", ct))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	// Validate the secret token by looking up connections with the provided secret token.
	cids, err := h.findAuthenticatedConnections(r.Context(), r, l)
	if err != nil {
		l.Error("incoming event: secret token validation failed", zap.Error(err))
		common.HTTPError(w, http.StatusUnauthorized)
		return nil
	}

	if len(cids) == 0 {
		// No connections found with the provided secret token.
		l.Warn("incoming event: no connections found with the provided secret token")
		common.HTTPError(w, http.StatusServiceUnavailable)
		return nil
	}

	// Read the request's JSON body, up to 8 MiB, to prevent DDoS attacks.
	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, 1<<23))
	if err != nil {
		l.Error("incoming event: failed to read HTTP body", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Warn("failed to parse Telegram update", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	return payload
}

// findAuthenticatedConnections validates the webhook request against configured secret tokens.
// Returns nil if validation passes, error otherwise.
func (h handler) findAuthenticatedConnections(ctx context.Context, r *http.Request, l *zap.Logger) ([]sdktypes.ConnectionID, error) {
	requestToken := r.Header.Get(headerSecretToken)

	// Find all Telegram bot connections
	cids, err := h.vars.FindConnectionIDs(ctx, desc.ID(), SecretToken, requestToken)
	if err != nil {
		l.Error("failed to find telegram connections", zap.Error(err))
		return nil, fmt.Errorf("failed to lookup connections: %w", err)
	}

	return cids, nil
}

// getTelegramEventType determines the Telegram event type based on the keys present in the event map.
func getTelegramEventType(event map[string]any) string {
	for _, eventType := range telegramEventTypes {
		if _, exists := event[eventType]; exists {
			return eventType
		}
	}

	return "unknown"
}
