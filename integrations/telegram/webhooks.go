package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleEvent processes incoming Telegram webhook events.
// Must respond with HTTP status 2XX promptly to indicate successful processing.
// If the response is not 2XX or times out (~60s), Telegram will retry the update
// multiple times until it eventually gives up after a reasonable number of attempts.
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	l := h.logger.With(
		zap.String("url_path", r.URL.Path),
		zap.String("event_type", r.Header.Get("Telegram-Event")),
	)

	botID := r.PathValue("connection_id")
	if botID == "" {
		l.Warn("missing bot ID in URL path")
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Find connection IDs for this webhook.
	ctx := r.Context()
	cids, err := h.vars.FindActiveConnectionIDs(ctx, desc.ID(), BotIDVar, botID)
	if err != nil {
		l.Error("failed to find connection IDs", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	if len(cids) == 0 {
		l.Warn("no connections found for bot ID", zap.String("bot_id", botID))
		return
	}

	// Check the request's headers, validate secret token, and parse body.
	telegramEvent := h.checkRequest(w, r, cids[0], l)
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

	// Dispatch the event to all connections for this webhook.
	common.DispatchEvent(ctx, h.logger, h.dispatch, akEvent, cids)
}

// checkRequest checks that the HTTP request has the right content type, validates the secret token,
// and parses the JSON body into a map.
func (h handler) checkRequest(w http.ResponseWriter, r *http.Request, connectionID sdktypes.ConnectionID, l *zap.Logger) map[string]any {
	// Check the request's HTTP headers.
	if common.PostWithoutJSONContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("incoming event: unexpected content type", zap.String("content_type", ct))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	if err := h.validateSecretTokenForConnection(r.Context(), r, connectionID, l); err != nil {
		l.Error("secret token validation failed", zap.Error(err))
		common.HTTPError(w, http.StatusUnauthorized)
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

// validateSecretTokenForConnection validates the secret token for a specific connection.
func (h handler) validateSecretTokenForConnection(ctx context.Context, r *http.Request, cid sdktypes.ConnectionID, l *zap.Logger) error {
	requestToken := r.Header.Get(headerSecretToken)
	if requestToken == "" {
		l.Warn("webhook missing required secret token",
			zap.String("connection_id", cid.String()))
		return errors.New("missing required secret token")
	}

	vs, errStatus, err := common.ReadVarsWithStatus(ctx, h.vars, cid)
	if errStatus.IsValid() || err != nil {
		l.Error("failed to read connection vars",
			zap.String("connection_id", cid.String()), zap.Error(err))
		return errors.New("failed to read connection vars")
	}
	webhookSecret := vs.GetValue(SecretTokenVar)

	if webhookSecret == "" {
		l.Warn("webhook has no secret token configured.",
			zap.String("connection_id", cid.String()))
		return errors.New("no secret token configured for connection")
	}

	if requestToken != webhookSecret {
		l.Warn("webhook secret token validation failed",
			zap.String("connection_id", cid.String()))
		return errors.New("invalid secret token")
	}

	return nil
}

// getTelegramEventType determines the Telegram event type based on the keys present in the event map.
func getTelegramEventType(event map[string]any) string {
	for _, eventType := range telegramEventTypes {
		if _, exists := event[eventType]; exists {
			return eventType
		}
	}

	return "unknown telegram event"
}
