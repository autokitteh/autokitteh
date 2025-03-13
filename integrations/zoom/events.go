package zoom

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

// handleEvent receives and dispatches asynchronous Zoom events.
// We must respond with 200 OK promptly (within 3 seconds).
// If the response takes longer than 3 seconds, Zoom will
// consider the delivery failed. Zoom will retry failed webhook deliveries up to 5 times
// with a progressively increasing delay (5s, 10s, 20s, 40s, 80s) between each retry.
// (Based on: https://developers.zoom.us/docs/api/webhooks/)
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check the request's headers and parse its body.
	zoomEvent := h.checkRequest(w, r)
	if zoomEvent == nil {
		return
	}

	eventType, ok := zoomEvent["event"].(string)
	if !ok || eventType == "" {
		h.logger.Warn("received Zoom event without event type", zap.Any("event", zoomEvent))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Extract account ID from the Zoom event payload.
	payload, ok := zoomEvent["payload"].(map[string]any)
	if !ok {
		h.logger.Warn("received Zoom event with invalid payload structure",
			zap.String("event_type", eventType),
			zap.Any("event", zoomEvent))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	accountID, ok := payload["account_id"].(string)
	if !ok {
		h.logger.Warn("received Zoom event without account ID",
			zap.String("event_type", eventType),
			zap.Any("event", zoomEvent))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Transform the Zoom event into an AutoKitteh event.
	akEvent, err := common.TransformEvent(h.logger, zoomEvent, eventType)
	if err != nil {
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := r.Context()
	cids, err := h.vars.FindConnectionIDs(ctx, desc.ID(), accountIDVar, accountID)
	if err != nil {
		h.logger.Error("failed to find connection IDs",
			zap.String("account_id", accountID),
			zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all connections for potential asynchronous handling.
	common.DispatchEvent(ctx, h.logger, h.dispatch, akEvent, cids)
}

// checkRequest validates the Zoom webhook HTTP request by verifying the content type,
// required headers, and signature and returns the parsed JSON payload if valid.
// Otherwise it returns nil with an HTTP error response.
func (h handler) checkRequest(w http.ResponseWriter, r *http.Request) map[string]any {
	l := h.logger.With(zap.String("url_path", r.URL.Path))

	// Check the request's HTTP headers.
	if common.PostWithoutJSONContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("incoming event: unexpected content type", zap.String("content_type", ct))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	signature := r.Header.Get("x-zm-signature")
	timestamp := r.Header.Get("x-zm-request-timestamp")

	if signature == "" || timestamp == "" {
		l.Warn("incoming event: missing Zoom verification headers")
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	// Read the request's JSON body, up to 8 MiB, to prevent DDoS attacks.
	body, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, 1<<23))
	if err != nil {
		l.Error("incoming event: failed to read request body", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return nil
	}

	secret, err := h.signingSecret()
	if err != nil {
		l.Error("incoming event: signing secret not found", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return nil
	}

	if secret == "" {
		// Zoom is not configured, so there's no point
		// in verifying or accepting the payload.
		return nil
	}

	// Verify Zoom signature.
	if !checkSignature(signature, timestamp, secret, body) {
		l.Warn("incoming event: invalid Zoom signature")
		common.HTTPError(w, http.StatusUnauthorized)
		return nil
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Warn("incoming event: failed to parse JSON body", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	l.Warn("incoming event: missing Zoom verification headers")

	return payload
}

// signingSecret retrieves the Zoom secret token from environment variables.
// TODO(INT-332): Support all AutoKitteh connection auth types.
func (h handler) signingSecret() (string, error) {
	secret := os.Getenv("ZOOM_SECRET_TOKEN")
	if secret == "" {
		return "", errors.New("missing Zoom secret token")
	}
	return secret, nil
}

// checkSignature compares the computed hash of the request body, using a preconfigured
// Zoom webhook secret token, to the signature provided in the request's `x-zm-signature`
// header, using HMAC-SHA256 for integrity verification.
// (based on https://developers.zoom.us/docs/api/webhooks/#verify-with-zooms-header)
func checkSignature(signature, timestamp, secret string, body []byte) bool {
	// Create HMAC SHA-256 hash using the webhook secret token.
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("v0:%s:%s", timestamp, string(body))))
	expectedSignature := "v0=" + hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
