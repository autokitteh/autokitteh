package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
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
	secrets       sdkservices.Secrets
	dispatcher    sdkservices.Dispatcher
	integrationID sdktypes.IntegrationID
	scope         string
}

func NewHandler(l *zap.Logger, sec sdkservices.Secrets, d sdkservices.Dispatcher, scope string, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		secrets:       sec,
		dispatcher:    d,
		scope:         scope,
		integrationID: id,
	}
}

// checkRequest checks that the given HTTP request has a valid content type and
// a valid Slack signature, and if so it returns the request's body. Otherwise
// it returns nil, and sends an HTTP error to the Slack platform's client.
func checkRequest(w http.ResponseWriter, r *http.Request, l *zap.Logger, wantContentType string) []byte {
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
	b, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("Failed to read inbound HTTP request body",
			zap.Error(err),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}
	signingSecret := os.Getenv(signingSecretEnvVar)
	if !verifySignature(signingSecret, ts, sig, b) {
		l.Error("Slack signature verification failed")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return nil
	}
	return b
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

// listTokens calls the List method in SecretsService.
func (h handler) listTokens(appID, enterpriseID, teamID string) ([]string, error) {
	k := appSecretName(appID, enterpriseID, teamID)
	tokens, err := h.secrets.List(context.Background(), h.scope, k)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func appSecretName(appID, enterpriseID, teamID string) string {
	s := fmt.Sprintf("apps/%s/%s/%s", appID, enterpriseID, teamID)
	// Slack enterprise ID is allowed to be empty.
	return strings.ReplaceAll(s, "//", "/")
}

func (h handler) dispatchAsyncEventsToConnections(l *zap.Logger, tokens []string, event *eventsv1.Event) {
	ctx := extrazap.AttachLoggerToContext(l, context.Background())
	for _, connToken := range tokens {
		event.IntegrationToken = connToken
		event := kittehs.Must1(sdktypes.EventFromProto(event))
		eventID, err := h.dispatcher.Dispatch(ctx, event, nil)
		if err != nil {
			l.Error("Dispatch failed",
				zap.String("connectionToken", connToken),
				zap.Error(err),
			)
			return
		}
		l.Debug("Dispatched",
			zap.String("connectionToken", connToken),
			zap.String("eventID", eventID.String()),
		)
	}
}
