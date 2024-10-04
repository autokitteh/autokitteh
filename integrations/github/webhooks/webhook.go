package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
		webhookID := r.PathValue("id")
		l := l.With(zap.String("webhookID", webhookID))

		ctx := r.Context()
		cids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, vars.PATKey, webhookID)
		if err != nil {
			l.Error("Failed to find connection IDs", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		for _, cid := range cids {
			l := l.With(zap.String("connectionID", cid.String()))

			data, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.PATSecret)
			if err != nil {
				l.Error("Unrecognized connection for user event payload", zap.Error(err))
				http.Error(w, "Internal Server error", http.StatusInternalServerError)
				return
			}

			payload, err = github.ValidatePayload(r, []byte(data.GetValue(vars.PATSecret)))
			if err != nil {
				l.Warn("Received invalid user event payload", zap.Error(err))
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if payload != nil {
				userCID = cid
				break // Successful validation with non-empty payload - no need to repeat.
			}
		}

		// The loop above tries to validate and parse the event, by finding the
		// AK connection corresponding to the webhook ID and using its secret.
		// If the payload is still nil at this point, then either the event is
		// fake (not from GitHub), or a relevant connection could not be found.
		// Either way, we report success (HTTP 200) and do nothing.
		if payload == nil {
			l.Info("Received GitHub event from user webhook, but no relevant connection found")
			return
		}
	}

	eventType := github.WebHookType(r)
	l = l.With(zap.String("eventType", eventType))

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
	// ghEvent is used to validate the payload and extract the installation ID.
	// To preserve the original event format instead of using the go-github event format,
	// we unmarshal the payload into a map and use that map to transform the event.
	var jsonEvent map[string]any
	err = json.Unmarshal(payload, &jsonEvent)
	if err != nil {
		l.Error("Failed to unmarshal payload to map", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Transform the received GitHub event into an AutoKitteh event.
	data, err := transformEvent(l, w, jsonEvent)
	if err != nil {
		return
	}
	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      data,
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.Any("data", data),
			zap.Error(err),
		)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Retrieve all the relevant connections for this event.
	var cids []sdktypes.ConnectionID
	if userCID.IsValid() {
		cids = append(cids, userCID) // User-defined webhook.
	}

	ctx := extrazap.AttachLoggerToContext(l, r.Context())

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
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)

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
func transformEvent(l *zap.Logger, w http.ResponseWriter, event any) (map[string]*sdktypes.ValuePB, error) {
	wrapped, err := sdktypes.WrapValue(event)
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

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	for _, cid := range cids {
		eid, err := h.dispatcher.Dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
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
