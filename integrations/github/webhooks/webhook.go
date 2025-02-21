package webhooks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
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

// ServeHTTP dispatches to autokitteh an asynchronous event notification that our GitHub app subscribed
// to. See https://github.com/organizations/autokitteh/settings/apps/autokitteh/permissions.
func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("url_path", r.URL.Path))
	var (
		payload []byte
		err     error
		userCID sdktypes.ConnectionID
	)
	if strings.HasSuffix(r.URL.Path, "/github/webhook") {
		// Validate that the inbound HTTP request has a valid content type
		// and a valid signature header, and if so parse the received event.
		secret, err := h.webhookSecret(r)
		if err != nil {
			l.Error("failed to get GitHub webhook secret", zap.Error(err))
			common.HTTPError(w, http.StatusInternalServerError)
			return
		}
		if secret == "" {
			// GitHub is not configured, so there's no point
			// in validating or accepting the payload.
			return
		}

		payload, err = github.ValidatePayload(r, []byte(secret))
		if err != nil {
			l.Warn("received invalid app event payload",
				zap.String("app_id", r.Header.Get(githubAppIDHeader)),
				zap.Error(err),
			)
			common.HTTPError(w, http.StatusForbidden)
			return
		}
	} else {
		// This event is for a user-defined webhook, not an app.
		webhookID := r.PathValue("id")
		l := l.With(zap.String("webhook_id", webhookID))

		ctx := r.Context()
		cids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, vars.PATKey, webhookID)
		if err != nil {
			l.Error("failed to find connection IDs", zap.Error(err))
			common.HTTPError(w, http.StatusInternalServerError)
			return
		}

		for _, cid := range cids {
			l := l.With(zap.String("connection_id", cid.String()))

			data, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.PATSecret)
			if err != nil {
				l.Error("unrecognized connection for user event payload", zap.Error(err))
				common.HTTPError(w, http.StatusInternalServerError)
				return
			}

			payload, err = github.ValidatePayload(r, []byte(data.GetValue(vars.PATSecret)))
			if err != nil {
				l.Warn("received invalid user event payload", zap.Error(err))
				common.HTTPError(w, http.StatusBadRequest)
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
			l.Info("received GitHub event from user webhook, but no relevant connection found")
			return
		}
	}

	eventType := github.WebHookType(r)
	l = l.With(zap.String("event_type", eventType))

	ghEvent, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		l.Warn("received unrecognized event type",
			zap.ByteString("payload", payload),
			zap.Error(err),
		)
		common.HTTPError(w, http.StatusNotImplemented)
		return
	}

	appID := r.Header.Get(githubAppIDHeader)

	var installID string
	if !userCID.IsValid() {
		if installID, err = extractInstallationID(ghEvent); err != nil {
			l.Error("failed to extract installation ID and user", zap.Error(err))
			common.HTTPError(w, http.StatusBadRequest)
			return
		}
	}
	// ghEvent is used to validate the payload and extract the installation ID.
	// To preserve the original event format instead of using the go-github event format,
	// we unmarshal the payload into a map and use that map to transform the event.
	var jsonEvent map[string]any
	err = json.Unmarshal(payload, &jsonEvent)
	if err != nil {
		l.Error("failed to unmarshal payload to map", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
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
		l.Error("failed to convert protocol buffer to SDK event",
			zap.Any("data", data),
			zap.Error(err),
		)
		common.HTTPError(w, http.StatusInternalServerError)
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
			l.Error("failed to find connection IDs", zap.Error(err))
			common.HTTPError(w, http.StatusInternalServerError)
			return
		}

		cids = append(cids, icids...)
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)

	// Returning immediately without an error = acknowledgement of receipt.
}

// webhookSecret reads the webhook secret from the private connection's
// variable, or uses the webhook secret of the server's default GitHub app.
func (h handler) webhookSecret(r *http.Request) (string, error) {
	appID := r.Header.Get(githubAppIDHeader)
	if appID == "" {
		return "", errors.New("missing GitHub app ID in event's HTTP header")
	}

	ctx := r.Context()
	cids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, vars.AppID, appID)
	if err != nil {
		return "", fmt.Errorf("failed to find connection IDs: %w", err)
	}
	if len(cids) == 0 {
		return "", nil
	}

	cid := cids[0] // Any connection will do, as they all share the same secret.
	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid), vars.ClientSecret)
	if err != nil {
		return "", fmt.Errorf("failed to read connection var: %w", err)
	}

	secret := vs.GetValue(vars.ClientSecret)
	if secret == "" {
		secret = os.Getenv(webhookSecretEnvVar)
	}

	return secret, nil
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
		l.Error("failed to wrap GitHub event", zap.Any("event", event), zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return nil, err
	}
	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("failed to convert wrapped GitHub event", zap.Any("event", event), zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return nil, err
	}
	return kittehs.TransformMapValues(data, sdktypes.ToProto), nil
}

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) {
	l := extrazap.ExtractLoggerFromContext(ctx)
	for _, cid := range cids {
		eid, err := h.dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		l := l.With(
			zap.String("connection_id", cid.String()),
			zap.String("event_id", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return
		}
		l.Debug("Event dispatched")
	}
}
