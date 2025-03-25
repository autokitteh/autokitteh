package jira

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger   *zap.Logger
	oauth    sdkservices.OAuth
	vars     sdkservices.Vars
	dispatch sdkservices.DispatchFunc
}

func NewHTTPHandler(l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars, d sdkservices.DispatchFunc) handler {
	l = l.With(zap.String("integration", desc.UniqueName().String()))
	return handler{logger: l, oauth: o, vars: v, dispatch: d}
}

// handleEvent receives asynchronous events from Jira,
// and dispatches them to zero or more AutoKitteh connections.
// Note 1: By default, AutoKitteh creates webhooks automatically,
// subscribing to all events - see "webhooks.go" for more details.
// Note 2: The requests are sent by a service, so no need to respond
// with user-friendly error web pages.
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("url_path", r.URL.Path))

	// Check the "Content-Type" header.
	if common.PostWithoutJSONContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("incoming event: unexpected content type", zap.String("content_type", ct))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Verify the JWT in the event's "Authorization" header.
	token := r.Header.Get(common.HeaderAuthorization)
	if !verifyJWT(l, strings.TrimPrefix(token, "Bearer ")) {
		l.Warn("Incoming Jira event with bad Authorization header")
		common.HTTPError(w, http.StatusUnauthorized)
		return
	}

	// Parse some of the metadata in the Jira event's JSON content.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Warn("Failed to read content of incoming Jira event", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	var jiraEvent map[string]any
	if err := json.Unmarshal(body, &jiraEvent); err != nil {
		l.Warn("Failed to unmarshal JSON in incoming Jira event", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	// Construct an AutoKitteh event from the Jira event.
	akEvent, err := constructEvent(l, jiraEvent)
	if err != nil {
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Iterate through all the relevant connections for this event.
	is, ok := jiraEvent["matchedWebhookIds"].([]any)
	if !ok {
		l.Warn("Invalid webhook IDs in Jira event", zap.Any("ids", jiraEvent["matchedWebhookIds"]))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	ids := kittehs.Transform(is, func(v any) int {
		f, ok := v.(float64)
		if !ok {
			l.Warn("Invalid webhook ID in Jira event", zap.Any("id", v))
			return 0
		}
		return int(f)
	})

	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	for _, id := range ids {
		u, err := transformIssueURL(jiraEvent, l)
		if err != nil {
			l.Error("Failed to transform Jira URL",
				zap.Error(err),
				zap.Any("jiraEvent", jiraEvent),
			)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		wk := webhookKey(u, strconv.Itoa(id))
		cids, err := h.vars.FindConnectionIDs(ctx, IntegrationID, WebhookKeySymbol, wk)
		if err != nil {
			l.Error("Failed to find connection IDs", zap.Error(err))
			break
		}

		common.DispatchEvent(ctx, l, h.dispatch, akEvent, cids)
	}

	// Returning immediately without an error = acknowledgement of receipt.
}

// transformIssueURL is used to transform the issue URL in the Jira event
// to a string that can be used for the webhook key.
func transformIssueURL(jiraEvent map[string]any, l *zap.Logger) (string, error) {
	issue, ok := jiraEvent["issue"].(map[string]any)
	if !ok {
		l.Warn("Invalid issue data in Jira event")
		return "", errors.New("invalid issue data")
	}

	selfURL, ok := issue["self"].(string)
	if !ok {
		l.Warn("Invalid issue URL in Jira event")
		return "", errors.New("invalid issue URL")
	}

	u, err := kittehs.NormalizeURL(selfURL, true)
	if err != nil {
		l.Warn("Invalid issue URL in Jira event", zap.Error(err))
		return "", fmt.Errorf("normalize URL: %w", err)
	}

	// Extract domain from the normalized URL
	domain, err := extractDomain(u)
	if err != nil {
		l.Warn("Invalid issue URL in Jira event", zap.Error(err))
		return "", fmt.Errorf("extract domain: %w", err)
	}

	return domain, nil
}

// extractDomain gets the subdomain from a Jira URL.
// For example, from "https://example.atlassian.net/...", returns "example".
func extractDomain(u string) (string, error) {
	parts := strings.Split(u, ".")
	if len(parts) < 2 {
		return "", errors.New("invalid URL format")
	}
	return path.Base(parts[0]), nil
}

// https://developer.atlassian.com/cloud/jira/platform/understanding-jwt-for-connect-apps/
func verifyJWT(l *zap.Logger, authz string) bool {
	token, err := jwt.Parse(authz, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			l.Warn("Unexpected signing method", zap.Any("alg", token.Header["alg"]))
		}
		// TODO(ENG-965): From new-connection form instead of env vars.
		return []byte(os.Getenv("JIRA_CLIENT_SECRET")), nil
	})
	if err != nil {
		l.Warn("Failed to parse JWT", zap.Error(err))
		return false
	}

	return token.Valid
}

func constructEvent(l *zap.Logger, jiraEvent map[string]any) (sdktypes.Event, error) {
	l = l.With(zap.Any("event", jiraEvent))

	wrapped, err := sdktypes.WrapValue(jiraEvent)
	if err != nil {
		l.Error("Failed to wrap Jira event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Jira event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	eventType, ok := jiraEvent["webhookEvent"].(string)
	if !ok {
		l.Error("Invalid event type")
		return sdktypes.InvalidEvent, errors.New("invalid event type")
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: strings.TrimPrefix(eventType, "jira:"),
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.String("eventType", strings.TrimPrefix(eventType, "jira:")),
			zap.Any("data", data),
			zap.Error(err),
		)
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}
