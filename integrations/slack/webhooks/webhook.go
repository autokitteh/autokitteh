package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/events"
	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// signingSecretEnvVar is the name of an environment variable that contains
	// a Slack app SECRET which is required to verify inbound request signatures.
	signingSecretEnvVar = "SLACK_SIGNING_SECRET"

	// The maximum shift/delay that we allow between an inbound request's
	// timestamp, and our current timestamp.
	maxDifference = 5 * time.Minute

	// Slack API implementation detail.
	slackSigVersion = "v0"
)

// handler is a collection of autokitteh webhooks which implement [http.Handler]
// to receive, dispatch, and acknowledge asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatch      sdkservices.DispatchFunc
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, vars sdkservices.Vars, d sdkservices.DispatchFunc, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		vars:          vars,
		dispatch:      d,
		integrationID: id,
	}
}

// checkRequest checks that the given HTTP request has a valid content type and
// a valid Slack signature, and if so it returns the request's body. Otherwise
// it returns nil, and sends an HTTP error to the Slack platform's client.
func (h handler) checkRequest(w http.ResponseWriter, r *http.Request, l *zap.Logger, wantContentType string) []byte {
	// "Content-Type" header.
	gotContentType := r.Header.Get(api.HeaderContentType)
	if gotContentType == "" || gotContentType != wantContentType {
		l.Error("Unexpected header value",
			zap.String("header", api.HeaderContentType),
			zap.String("got", gotContentType),
			zap.String("want", wantContentType),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}

	// "X-Slack-Request-Timestamp" header.
	ts := r.Header.Get(api.HeaderSlackTimestamp)
	if ts == "" {
		l.Warn("Missing header",
			zap.String("header", api.HeaderSlackTimestamp),
		)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}
	secs, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		l.Warn("Invalid header value",
			zap.String("header", api.HeaderSlackTimestamp),
			zap.String("value", ts),
		)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}
	d := time.Since(time.Unix(secs, 0))
	if d.Abs() > maxDifference {
		l.Warn("Unacceptable header value",
			zap.String("header", api.HeaderSlackTimestamp),
			zap.String("difference", fmt.Sprint(d)),
		)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}

	// "X-Slack-Signature" header.
	sig := r.Header.Get(api.HeaderSlackSignature)
	if sig == "" {
		l.Warn("Missing header",
			zap.String("header", api.HeaderSlackSignature),
		)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return nil
	}

	// Request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("Failed to read inbound HTTP request body",
			zap.Error(err),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}

	appID, teamID, enterpriseID, err := h.extractIDs(body, wantContentType, l)
	if err != nil {
		l.Error("Failed to extract IDs", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}

	// Get signing secret.
	cids, err := h.listConnectionIDs(r.Context(), appID, enterpriseID, teamID)
	if err != nil {
		l.Error("Failed to list connection IDs", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil
	}
	// No connections = respond to Slack with 200 OK, but don't process the request.
	if len(cids) == 0 {
		return nil
	}
	// All connections for the same app/enterprise/workspace share the same signing secret.
	secret, err := h.vars.Get(r.Context(), sdktypes.NewVarScopeID(cids[0]))
	if err != nil {
		l.Error("Failed to get signing secret", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil
	}
	signingSecret := secret.GetValue(vars.SigningSecret)
	if signingSecret == "" {
		signingSecret = os.Getenv(signingSecretEnvVar)
	}

	// Verify signature.
	if !verifySignature(signingSecret, ts, sig, body) {
		l.Error("Signature verification failed",
			zap.String("app_id", appID),
			zap.String("team_id", teamID),
		)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}

	return body
}

// verifySignature implements https://api.slack.com/authentication/verifying-requests-from-slack.
func verifySignature(signingSecret, ts, want string, body []byte) bool {
	mac := hmac.New(sha256.New, []byte(signingSecret))

	n, err := mac.Write([]byte(fmt.Sprintf("%s:%s:", slackSigVersion, ts)))
	if err != nil || n != len(ts)+4 {
		return false
	}

	if n, err := mac.Write(body); err != nil || n != len(body) {
		return false
	}

	got := fmt.Sprintf("%s=%s", slackSigVersion, hex.EncodeToString(mac.Sum(nil)))
	return hmac.Equal([]byte(got), []byte(want))
}

// Transform the received Slack event into an AutoKitteh event.
func transformEvent(l *zap.Logger, slackEvent any, eventType string) (sdktypes.Event, error) {
	l = l.With(
		zap.String("eventType", eventType),
		zap.Any("event", slackEvent),
	)

	wrapped, err := sdktypes.WrapValue(slackEvent)
	if err != nil {
		l.Error("Failed to wrap Slack event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Slack event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.Any("data", data),
			zap.Error(err),
		)
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}

func (h handler) listConnectionIDs(ctx context.Context, appID, enterpriseID, teamID string) ([]sdktypes.ConnectionID, error) {
	key := vars.KeyValue(appID, enterpriseID, teamID)
	return h.vars.FindConnectionIDs(ctx, h.integrationID, vars.KeyName, key)
}

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	for _, cid := range cids {
		eid, err := h.dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		l := l.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("Event dispatched")
	}
}

func (h handler) extractIDs(body []byte, wantContentType string, l *zap.Logger) (appID, teamID, enterpriseID string, err error) {
	// Option 1: JSON payloads.
	if wantContentType == "application/json" {
		var cb events.Callback
		if err := json.Unmarshal(body, &cb); err != nil {
			l.Error("Failed to parse JSON for app/team IDs",
				zap.Error(err),
			)
			return "", "", "", err
		}
		return cb.APIAppID, cb.TeamID, "", nil
	}

	// Option 2: URL-encoded web form payloads.
	kv, err := url.ParseQuery(string(body))
	if err != nil {
		l.Error("Failed to parse URL-encoded form",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		return "", "", "", err
	}

	// Check if this is an interaction payload.
	if payload := kv.Get("payload"); payload != "" {
		var p BlockActionsPayload
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			l.Error("Failed to parse interaction payload",
				zap.String("payload", payload),
				zap.Error(err),
			)
			return "", "", "", err
		}

		if p.IsEnterpriseInstall && p.Enterprise != nil {
			return p.APIAppID, p.Team.ID, p.Enterprise.ID, nil
		}
		return p.APIAppID, p.Team.ID, "", nil
	}

	// Regular form data (bot events and slash commands).
	return kv.Get("api_app_id"), kv.Get("team_id"), "", nil
}
