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
// Based on: https://developers.linear.app/docs/graphql/webhooks
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	payload := h.checkRequest(w, r)
	if payload == nil {
		return
	}

	// TODO: Transform the received Linear event into an AutoKitteh event.

	// TODO: Retrieve all the relevant connections for this event.

	// TODO: Dispatch the event to all of them, for asynchronous handling.
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
	ct := r.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
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
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<23))
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

	if !checkSignature(secret, sig, body) {
		l.Warn("incoming event: signature check failed", zap.String("signature", sig))
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

// TODO(INT-284): Support all AutoKitteh connection auth types.
func (h handler) signingSecret() (string, error) {
	secret := os.Getenv("LINEAR_SIGNING_SECRET")
	if secret == "" {
		return "", errors.New("missing Linear signing secret")
	}
	return secret, nil
}

// checkSignature compares the event body's hash, using a preconfigured signing
// secret, to a signature reported in the event's HTTP header, using SHA256 HMAC.
// Based on: https://developers.linear.app/docs/graphql/webhooks#securing-webhooks
func checkSignature(secret, want string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(secret))

	if n, err := mac.Write(body); err != nil || n != len(body) {
		return false
	}

	got := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(got), []byte(want))
}

// checkTimestamp checks that the event is fresh, to prevent replay attacks.
// Based on: https://developers.linear.app/docs/graphql/webhooks#securing-webhooks
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
