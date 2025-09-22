package hubspot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	portalIDVar      = sdktypes.NewSymbol("portal_id")
	allowed_Duration = 5 * time.Minute // Default allowed duration for HubSpot signature.
)

// handleEvent receives and dispatches asynchronous HubSpot events.
// We must respond with 200 OK within 5 seconds, otherwise HubSpot
// will consider the event failed and may retry.
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Check the request's headers and parse its body.
	hubspotEvents := h.checkRequest(w, r)
	if hubspotEvents == nil {
		return
	}

	eventType, ok := hubspotEvents["subscriptionType"].(string)
	if !ok {
		h.logger.Warn("received HubSpot event without subscription type",
			zap.Any("event", hubspotEvents),
		)
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Transform the HubSpot event into an AutoKitteh event.
	akEvent, err := common.TransformEvent(h.logger, hubspotEvents, eventType)
	if err != nil {
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	var portalID string
	if id, ok := hubspotEvents["portalId"].(float64); ok {
		portalID = strconv.FormatFloat(id, 'f', 0, 64)
	} else {
		h.logger.Warn("received HubSpot event without portal ID",
			zap.String("event_type", eventType),
			zap.Any("event", hubspotEvents),
		)
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Retrieve all the relevant connections for this event.
	ctx := r.Context()
	cids, err := h.vars.FindActiveConnectionIDs(ctx, desc.ID(), portalIDVar, portalID)
	if err != nil {
		h.logger.Error("failed to find connection IDs", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Dispatch the event to all connections for potential asynchronous handling.
	common.DispatchEvent(ctx, h.logger, h.dispatch, akEvent, cids)
}

func (h handler) checkRequest(w http.ResponseWriter, r *http.Request) map[string]any {
	l := h.logger.With(zap.String("url_path", r.URL.Path))

	// Check the request's HTTP headers.
	if common.PostWithoutJSONContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("incoming event: unexpected content type", zap.String("content_type", ct))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	signature := r.Header.Get("X-HubSpot-Signature-V3") // HubSpot version 3 signature.
	timestamp := r.Header.Get("X-HubSpot-Request-Timestamp")
	proto := r.Header.Get("X-Forwarded-Proto")

	if signature == "" || timestamp == "" {
		l.Warn("incoming event: missing HubSpot signature header")
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
		// HubSpot is not configured, so there's no point
		// in verifying or accepting the payload.
		return nil
	}

	// Construct full URL for v3 signature.
	uri := fmt.Sprintf("%s://%s%s", proto, r.Host, r.URL.RequestURI())

	valid, err := checkSignature(signature, timestamp, secret, body, r.Method, uri)
	if err != nil {
		l.Error("incoming event: signature check failed", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return nil
	}
	if !valid {
		l.Warn("incoming event: invalid HubSpot signature")
		common.HTTPError(w, http.StatusUnauthorized)
		return nil
	}

	var payload []map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		l.Warn("incoming event: failed to parse JSON body", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	return payload[0]
}

func (h handler) signingSecret() (string, error) {
	secret := os.Getenv("HUBSPOT_CLIENT_SECRET")
	if secret == "" {
		return "", errors.New("missing HubSpot secret token")
	}
	return secret, nil
}

func checkSignature(signature, timestamp, secret string, body []byte, method, uri string) (bool, error) {
	currentTime := time.Now()
	tsInt, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return false, fmt.Errorf("parse int(%s): %w", timestamp, err)
	}
	tsTime := time.UnixMilli(tsInt)

	// Check if the timestamp is within the allowed duration according to hubspot's guidelines.
	if currentTime.Sub(tsTime) > allowed_Duration {
		return false, fmt.Errorf("timestamp %s is older than allowed duration %s", tsTime, allowed_Duration)
	}

	// HubSpot v3 signature: method + uri + body + timestamp.
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%s%s%s%s", method, uri, string(body), timestamp)
	expectedSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature)), nil
}
