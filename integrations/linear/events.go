package linear

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
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

// handleEvent receives and dispatches asynchronous Linear events.
// We must respond with 200 OK within 5 seconds, otherwise Linear
// will retry up to 3 times (with exponential backoff: 1m, 1h, 6h).
// (Based on: https://developers.linear.app/docs/graphql/webhooks).
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check the request's headers and parse its body.
	linearEvent := h.checkRequest(w, r)
	if linearEvent == nil {
		return
	}

	eventType := strings.ToLower(r.Header.Get("Linear-Event"))
	orgID, ok := linearEvent["organizationId"].(string)
	if !ok {
		h.logger.Warn("received Linear event without organization ID",
			zap.String("event_type", eventType),
			zap.Any("event", linearEvent),
		)
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Transform the Linear event into an AutoKitteh event.
	akEvent, err := common.TransformEvent(h.logger, linearEvent, eventType)
	if err != nil {
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := r.Context()
	cids, err := h.vars.FindConnectionIDs(ctx, desc.ID(), orgIDVar, orgID)
	if err != nil {
		h.logger.Error("failed to find connection IDs", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all of them, for potential asynchronous handling.
	common.DispatchEvent(ctx, h.logger, h.dispatch, akEvent, cids)
}

// checkRequest checks that the HTTP request has the right content
// type and signature, and if so it returns the request's JSON payload.
// Otherwise it returns nil, and responds to the sender with an HTTP error.
func (h handler) checkRequest(w http.ResponseWriter, r *http.Request) map[string]any {
	l := h.logger.With(
		zap.String("url_path", r.URL.Path),
		zap.String("event_type", r.Header.Get("Linear-Event")),
	)

	// No need to check the HTTP method, as we only accept POST requests.

	// Check the request's HTTP headers.
	if common.PostWithoutJSONContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("incoming event: unexpected content type", zap.String("content_type", ct))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	sig := r.Header.Get("Linear-Signature")
	if sig == "" {
		l.Warn("incoming event: missing header", zap.String("header", "Linear-Signature"))
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

	// Check the event's signature to prevent impersonation to Linear.
	secret, err := h.signingSecret()
	if err != nil {
		l.Error("incoming event: signing secret not found", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return nil
	}

	if secret == "" {
		// Linear is not configured, so there's no point
		// in verifying or accepting the payload.
		return nil
	}

	if !checkSignature(l, secret, sig, body) {
		common.HTTPError(w, http.StatusUnauthorized)
		return nil
	}

	// Check the event's timestamp is within the past minute, to prevent replay attacks.
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Warn("incoming event: failed to parse JSON body", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	if err := checkTimestamp(payload); err != nil {
		l.Warn("incoming event: " + err.Error())
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	return payload
}

// TODO(INT-332): Support all AutoKitteh connection auth types.
func (h handler) signingSecret() (string, error) {
	secret := os.Getenv("LINEAR_WEBHOOK_SECRET")
	if secret == "" {
		return "", errors.New("missing Linear signing secret")
	}
	return secret, nil
}

// checkSignature compares the event body's hash, using a preconfigured signing
// secret, to a signature reported in the event's HTTP header, using SHA256 HMAC
// (based on: https://developers.linear.app/docs/graphql/webhooks#securing-webhooks).
func checkSignature(l *zap.Logger, secret, want string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(secret))

	if n, err := mac.Write(body); err != nil || n != len(body) {
		return false
	}

	got := hex.EncodeToString(mac.Sum(nil))
	match := hmac.Equal([]byte(got), []byte(want))
	if !match {
		l.Warn("incoming event: signature check failed",
			zap.String("expected_signature", want),
			zap.String("actual_signature", got),
		)
	}
	return match
}

// checkTimestamp checks that the event is fresh, to prevent replay attacks
// (based on: https://developers.linear.app/docs/graphql/webhooks#securing-webhooks).
func checkTimestamp(payload map[string]any) error {
	wts, ok := payload["webhookTimestamp"]
	if !ok {
		return errors.New("missing webhook timestamp")
	}

	msec, ok := wts.(float64)
	if !ok {
		return fmt.Errorf("invalid webhook timestamp: %v", wts)
	}

	t := time.UnixMilli(int64(msec))
	d := time.Since(t)
	if d < 0 {
		return errors.New("webhook timestamp in the future: " + t.Format(time.RFC3339))
	}
	if d > time.Minute {
		return errors.New("webhook timestamp is stale: " + t.Format(time.RFC3339))
	}

	return nil
}
