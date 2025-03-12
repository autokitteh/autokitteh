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
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/slack/events"
	"go.autokitteh.dev/autokitteh/integrations/slack/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerSlackTimestamp = "X-Slack-Request-Timestamp"
	headerSlackSignature = "X-Slack-Signature"

	// The maximum shift/delay that we allow between an inbound request's
	// timestamp, and our current timestamp, to defend against replay attacks.
	// See https://api.slack.com/authentication/verifying-requests-from-slack.
	maxDifference = 5 * time.Minute

	// Slack API implementation detail.
	slackSigVersion = "v0"
)

// handler implements several HTTP webhooks to receive and
// dispatch third-party asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatch      sdkservices.DispatchFunc
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.DispatchFunc, i sdktypes.IntegrationID) handler {
	return handler{logger: l, vars: v, dispatch: d, integrationID: i}
}

// checkRequest checks that the given HTTP request has a valid content type and
// a valid Slack signature, and if so it returns the request's body. Otherwise
// it returns nil, and sends an HTTP error to the Slack platform's client.
func (h handler) checkRequest(w http.ResponseWriter, r *http.Request, l *zap.Logger, wantContentType string) []byte {
	// "Content-Type" header.
	gotContentType := r.Header.Get(common.HeaderContentType)
	if gotContentType == "" || gotContentType != wantContentType {
		l.Warn("incoming event: unexpected header value",
			zap.String("header", common.HeaderContentType),
			zap.String("got", gotContentType),
			zap.String("want", wantContentType),
		)
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	// "X-Slack-Request-Timestamp" header.
	ts := r.Header.Get(headerSlackTimestamp)
	if ts == "" {
		l.Warn("missing header", zap.String("header", headerSlackTimestamp))
		common.HTTPError(w, http.StatusUnauthorized)
		return nil
	}
	secs, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		l.Warn("invalid header value",
			zap.String("header", headerSlackTimestamp),
			zap.String("value", ts),
		)
		common.HTTPError(w, http.StatusForbidden)
		return nil
	}
	d := time.Since(time.Unix(secs, 0))
	if d.Abs() > maxDifference {
		l.Warn("unacceptable header value",
			zap.String("header", headerSlackTimestamp),
			zap.String("difference", fmt.Sprint(d)),
		)
		common.HTTPError(w, http.StatusForbidden)
		return nil
	}

	// "X-Slack-Signature" header.
	sig := r.Header.Get(headerSlackSignature)
	if sig == "" {
		l.Warn("missing header", zap.String("header", headerSlackSignature))
		common.HTTPError(w, http.StatusUnauthorized)
		return nil
	}

	// Request body.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("failed to read inbound HTTP request body", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	// Verify signature.
	secret, info, err := h.signingSecret(r.Context(), l, body, wantContentType)
	if err != nil {
		l.Error("failed to get Slack signing secret", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return nil
	}
	if secret == "" {
		// Slack is not configured, so there's no point
		// in verifying or accepting the payload.
		return nil
	}

	if !verifySignature(secret, ts, sig, body) {
		l.Warn("signature verification failed",
			zap.String("signature", sig),
			zap.Bool("has_signing_secret", secret != ""),
			zap.String("app_id", info.AppID),
			zap.String("enterprise_id", info.EnterpriseID),
			zap.String("team_id", info.TeamID),
		)
		common.HTTPError(w, http.StatusForbidden)
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
		zap.String("event_type", eventType),
		zap.Any("event", slackEvent),
	)

	wrapped, err := sdktypes.WrapValue(slackEvent)
	if err != nil {
		l.Error("failed to wrap Slack event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("failed to convert wrapped Slack event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("failed to convert protocol buffer to SDK event",
			zap.Any("data", data),
			zap.Error(err),
		)
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}

func (h handler) listConnectionIDs(ctx context.Context, appID, enterpriseID, teamID string) ([]sdktypes.ConnectionID, error) {
	ids := vars.InstallIDs(appID, enterpriseID, teamID)
	return h.vars.FindConnectionIDs(ctx, h.integrationID, vars.InstallIDsVar, ids)
}

// extractIDs extracts the app ID, team ID, and enterprise ID from the given request body.
func (h handler) extractIDs(l *zap.Logger, body []byte, wantContentType string) (string, string, string, error) {
	// Option 1: JSON payloads.
	if strings.HasPrefix(wantContentType, "application/json") {
		var cb events.Callback
		if err := json.Unmarshal(body, &cb); err != nil {
			l.Warn("failed to parse JSON for app/team IDs", zap.Error(err))
			return "", "", "", err
		}
		// TODO: add Enterprise support.
		return cb.APIAppID, "", cb.TeamID, nil
	}

	// Option 2: URL-encoded web form payloads.
	kv, err := url.ParseQuery(string(body))
	if err != nil {
		l.Warn("failed to parse URL-encoded form for app/team IDs",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		return "", "", "", err
	}

	// Regular form data (bot events and slash commands).
	payload := kv.Get("payload")
	if payload == "" {
		// TODO: add Enterprise support.
		return kv.Get("api_app_id"), "", kv.Get("team_id"), nil
	}

	// Interaction payloads.
	var p BlockActionsPayload
	if err := json.Unmarshal([]byte(payload), &p); err != nil {
		l.Warn("failed to parse interaction payload for app/team IDs",
			zap.String("payload", payload),
			zap.Error(err),
		)
		return "", "", "", err
	}

	return p.APIAppID, p.Enterprise.ID, p.Team.ID, nil
}

// signingSecret reads the signing secret from the private connection's
// variable, or uses the signing secret of the server's default Slack app.
// It also returns the app ID, team ID, and enterprise ID of the connection
// (if there's a signature verification error they're useful for debugging).
func (h handler) signingSecret(ctx context.Context, l *zap.Logger, body []byte, wantContentType string) (string, *vars.InstallInfo, error) {
	appID, enterpriseID, teamID, err := h.extractIDs(l, body, wantContentType)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract IDs: %w", err)
	}

	cids, err := h.listConnectionIDs(ctx, appID, enterpriseID, teamID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to find connection IDs: %w", err)
	}
	if len(cids) == 0 {
		return "", nil, nil
	}

	cid := cids[0] // Any connection will do, as they all share the same secret.
	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.SigningSecretVar)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read connection var: %w", err)
	}

	secret := vs.GetValue(vars.SigningSecretVar)
	if secret == "" {
		secret = os.Getenv(vars.SigningSecretEnvVar)
	}

	info := &vars.InstallInfo{
		AppID:        appID,
		EnterpriseID: enterpriseID,
		TeamID:       teamID,
	}

	return secret, info, err
}
