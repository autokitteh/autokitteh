package webhooks

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	valuesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
	vars          sdkservices.Vars
	dispatcher    sdkservices.Dispatcher
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, vars sdkservices.Vars, d sdkservices.Dispatcher, id sdktypes.IntegrationID) handler {
	return handler{
		logger:        l,
		vars:          vars,
		dispatcher:    d,
		integrationID: id,
	}
}

// ServeHTTP dispatches to autokitteh an asynchronous event notification that our GitHub app subscribed
// to. See https://github.com/organizations/autokitteh/settings/apps/autokitteh/permissions.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))
	var (
		payload []byte
		err     error
		userCID sdktypes.ConnectionID
	)
	if strings.HasSuffix(r.URL.Path, "/github/webhook") {
		// Validate that the inbound HTTP request has a valid content type
		// and a valid signature header, and if so parse the received event.
		payload, err = github.ValidatePayload(r, []byte(os.Getenv(webhookSecretEnvVar)))
		if err != nil {
			l.Warn("Received invalid app event payload", zap.Error(err))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
	} else {
		// This event is for a user-defined webhook, not an app.
		path := strings.Split(r.URL.Path, "/")
		suffix := path[len(path)-1]
		ctx := r.Context()

		cids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, vars.PATKey, suffix)
		if err != nil {
			l.Error("Failed to find connection IDs", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for _, cid := range cids {
			data, err := h.vars.Reveal(ctx, sdktypes.NewVarScopeID(cid), vars.PATSecret)
			if err != nil {
				l.Warn("Unrecognized connection for user event payload",
					zap.String("suffix", suffix),
					zap.String("cid", cid.String()),
					zap.Error(err),
				)
				http.Error(w, "Internal Server error", http.StatusInternalServerError)
				return
			}

			payload, err = github.ValidatePayload(r, []byte(data.GetValue(vars.PATSecret)))
			if err != nil {
				l.Warn("Received invalid user event payload", zap.Error(err))
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		}
	}

	eventType := github.WebHookType(r)
	ghEvent, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		l.Warn("Received unrecognized event type",
			zap.ByteString("payload", payload),
			zap.Error(err),
		)
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
		return
	}

	appID := r.Header.Get(githubAppIDHeader)

	var installID string
	if !userCID.IsValid() {
		if installID, err = extractInstallationID(ghEvent); err != nil {
			l.Error("Failed to extract installation ID and user", zap.Error(err))
			http.Error(w, "no installation or user specified", http.StatusBadRequest)
			return
		}
	}

	// Transform the received GitHub event into an autokitteh event.
	data, err := transformEvent(l, w, ghEvent)
	if err != nil {
		return
	}
	akEvent := &sdktypes.EventPB{
		EventType: eventType,
		Data:      data,
	}

	// Retrieve all the relevant connections for this event.
	var cids []sdktypes.ConnectionID
	if userCID.IsValid() {
		cids = append(cids, userCID) // User-defined webhook.
	}

	ctx := r.Context()

	if installID != "" {
		// App webhook.
		icids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, vars.InstallKey(appID, installID), "")
		if err != nil {
			l.Error("Failed to find connection IDs", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		cids = append(cids, icids...)
	}

	// Dispatch the event to all of them, for asynchronous handling.
	dispatchAsyncEventsToConnections(ctx, l, cids, akEvent, h.dispatcher)

	// Returning immediately without an error = acknowledgement of receipt.
}

func extractInstallationID(event any) (inst string, err error) {
	type itf interface {
		GetInstallation() *github.Installation
	}

	obj, ok := event.(itf)
	if !ok {
		err = errors.New("event does not have installation")
		return
	}

	pinst := obj.GetInstallation()

	if pinst == nil {
		err = errors.New("event does not have installation")
		return
	}

	inst = strconv.FormatInt(*(pinst).ID, 10)
	return
}

// transformEvent transforms a received GitHub event into an autokitteh event.
func transformEvent(l *zap.Logger, w http.ResponseWriter, event any) (map[string]*valuesv1.Value, error) {
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(event)
	if err != nil {
		l.Error("Failed to wrap GitHub event", zap.Any("event", event), zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped GitHub event", zap.Any("event", event), zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return nil, err
	}
	return kittehs.TransformMapValues(data, sdktypes.ToProto), nil
}

func dispatchAsyncEventsToConnections(ctx context.Context, l *zap.Logger, cids []sdktypes.ConnectionID, event *eventsv1.Event, d sdkservices.Dispatcher) {
	for _, cid := range cids {
		event.ConnectionId = cid.String()
		e, err := sdktypes.EventFromProto(event)
		if err != nil {
			l.Error("Failed to convert protocol buffer to SDK event", zap.Error(err))
			return
		}

		eventID, err := d.Dispatch(ctx, e, nil)
		if err != nil {
			l.Error("Event dispatch failed",
				zap.String("eventID", eventID.String()),
				zap.String("connectionID", cid.String()),
				zap.Error(err),
			)
			return
		}
		l.Debug("Event dispatched",
			zap.String("eventID", eventID.String()),
			zap.String("connectionID", cid.String()),
		)
	}
}
