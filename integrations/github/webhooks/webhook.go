package webhooks

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/go-github/v54/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
)

const (
	// WebhookPath is the URL path for our webhook to handle inbound events.
	WebhookPath = "/github/webhook"

	// webhookSecretEnvVar is the name of an environment variable that contains a
	// GitHub app SECRET which is required to verify inbound request signatures.
	webhookSecretEnvVar = "GITHUB_WEBHOOK_SECRET"

	// githubAppIDHeader is the HTTP header that contains the GitHub app ID of an incoming event.
	githubAppIDHeader = "X-GitHub-Hook-Installation-Target-ID"
)

// handler is an autokitteh webhook which implements [http.Handler] to
// receive, dispatch, and acknowledge asynchronous event notifications.
type handler struct {
	logger        *zap.Logger
	secrets       sdkservices.Secrets
	dispatcher    sdkservices.Dispatcher
	scope         string
	integrationID sdktypes.IntegrationID
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

// ServeHTTP dispatches to autokitteh an asynchronous event notification that our GitHub app subscribed
// to. See https://github.com/organizations/autokitteh/settings/apps/autokitteh/permissions.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))
	var (
		payload       []byte
		err           error
		userConnToken string
	)
	if strings.HasSuffix(r.URL.Path, "/github/webhook") {
		// Validate that the inbound HTTP request has a valid content type
		// and a valid signature header, and if so parse the received event.
		payload, err = github.ValidatePayload(r, []byte(os.Getenv(webhookSecretEnvVar)))
		if err != nil {
			l.Warn("Received invalid app event payload",
				zap.Error(err),
			)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	} else {
		// This event is for a user-defined webhook, not an app.
		path := strings.Split(r.URL.Path, "/")
		suffix := path[len(path)-1]

		tokens, err := h.secrets.List(r.Context(), h.scope, "webhooks/"+suffix)
		if err != nil {
			l.Warn("Unrecognized user event payload",
				zap.Error(err),
			)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		if len(tokens) != 1 {
			l.Warn("Unexpected number of connection tokens for user event",
				zap.String("suffix", suffix),
				zap.Strings("tokens", tokens),
			)
			http.Error(w, "Internal Server error", http.StatusInternalServerError)
			return
		}
		userConnToken = tokens[0]
		data, err := h.secrets.Get(r.Context(), h.scope, userConnToken)
		if err != nil {
			l.Warn("Unrecognized connection for user event payload",
				zap.String("suffix", suffix),
				zap.String("token", userConnToken),
				zap.Error(err),
			)
			http.Error(w, "Internal Server error", http.StatusInternalServerError)
			return
		}

		payload, err = github.ValidatePayload(r, []byte(data["secret"]))
		if err != nil {
			l.Warn("Received invalid user event payload",
				zap.Error(err),
			)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	}

	eventType := github.WebHookType(r)
	ghEvent, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		l.Warn("Received unrecognized event type",
			zap.Error(err),
			zap.ByteString("payload", payload),
		)
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
		return
	}
	appID := r.Header.Get(githubAppIDHeader)
	installID := ""
	if userConnToken == "" {
		installID = extractInstallationID(l, ghEvent, eventType)
	}

	// Transform the received GitHub event into an autokitteh event.
	data, err := transformEvent(l, w, ghEvent)
	if err != nil {
		return
	}
	akEvent := &sdktypes.EventPB{
		IntegrationId:   h.integrationID.String(),
		OriginalEventId: github.DeliveryID(r),
		EventType:       eventType,
		Data:            data,
	}

	// Retrieve all the relevant connections for this event.
	connTokens := []string{userConnToken} // User-defined webhook.
	if installID != "" {
		// App webhook.
		connTokens, err = h.listTokens(appID, installID)
		if err != nil {
			l.Error("Failed to retrieve connection tokens",
				zap.Error(err),
			)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	}

	// Dispatch the event to all of them, for asynchronous handling.
	dispatchAsyncEventsToConnections(l, connTokens, akEvent, h.dispatcher)

	// Returning immediately without an error = acknowledgement of receipt.
}

func extractInstallationID(l *zap.Logger, event any, eventType string) string {
	v := reflect.Indirect(reflect.ValueOf(event)).FieldByName("Installation")
	if !v.IsValid() {
		l.Warn("Received event without installation details",
			zap.String("type", eventType),
			zap.Any("event", event),
		)
		return ""
	}
	id := *v.Elem().FieldByName("ID").Interface().(*int64)
	return strconv.FormatInt(id, 10)
}

// transformEvent transforms a received GitHub event into an autokitteh event.
func transformEvent(l *zap.Logger, w http.ResponseWriter, event any) (map[string]*valuesv1.Value, error) {
	wrapped, err := sdkvalues.DefaultValueWrapper.Wrap(event)
	if err != nil {
		l.Error("Failed to wrap GitHub event",
			zap.Error(err),
			zap.Any("event", event),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	data, err := sdktypes.ValueToStringValuesMap(wrapped)
	if err != nil {
		l.Error("Failed to convert wrapped GitHub event",
			zap.Error(err),
			zap.Any("event", event),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	return kittehs.TransformMapValues(data, sdktypes.ToProto), nil
}

// listTokens calls the List method in SecretsService.
// Applies only to GitHub app events, not user-defined webhooks.
func (h handler) listTokens(appID, installID string) ([]string, error) {
	key := fmt.Sprintf("apps/%s/%s", appID, installID)
	tokens, err := h.secrets.List(context.Background(), h.scope, key)
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func dispatchAsyncEventsToConnections(l *zap.Logger, tokens []string, event *eventsv1.Event, d sdkservices.Dispatcher) {
	ctx := extrazap.AttachLoggerToContext(l, context.Background())
	for _, connToken := range tokens {
		event.IntegrationToken = connToken
		e := kittehs.Must1(sdktypes.EventFromProto(event))
		eventID, err := d.Dispatch(ctx, e, nil)
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
