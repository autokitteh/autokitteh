package forms

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/api/forms/v1"

	"go.autokitteh.dev/autokitteh/integrations/google/internal/vars"
	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger     *zap.Logger
	vars       sdkservices.Vars
	dispatcher sdkservices.Dispatcher
}

func NewWebhookHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.Dispatcher) http.Handler {
	return handler{logger: l, vars: v, dispatcher: d}
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	formsEvent := map[string]any{
		"event_type":   r.Header.Get("Eventtype"),
		"form_id":      r.Header.Get("Formid"),
		"publish_time": r.Header.Get("X-Goog-Pubsub-Publish-Time"),
	}
	eventType := WatchEventType(r.Header.Get("Eventtype"))
	watchID := r.Header.Get("Watchid")

	l := h.logger.With(
		zap.String("urlPath", r.URL.Path),
		zap.String("eventType", r.Header.Get("Eventtype")),
		zap.String("formID", r.Header.Get("Formid")),
		zap.String("watchID", watchID),
		zap.String("messageID", r.Header.Get("X-Goog-Pubsub-Message-Id")),
		zap.String("publishTime", r.Header.Get("X-Goog-Pubsub-Publish-Time")),
		zap.String("subscriptionName", r.Header.Get("X-Goog-Pubsub-Subscription-Name")),
	)
	l.Info("Received Google Forms notification")

	name := vars.FormResponsesWatchID
	if eventType == WatchSchemaChanges {
		name = vars.FormSchemaWatchID
	}

	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	if err := h.dispatchAsyncEventsToConnections(ctx, name, watchID, formsEvent); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Returning immediately without an error = acknowledgement of receipt.
}

func (h handler) constructEvent(ctx context.Context, formsEvent map[string]any, cids []sdktypes.ConnectionID) (sdktypes.Event, error) {
	l := extrazap.ExtractLoggerFromContext(ctx)

	// Enrich the event with relevant data, with API calls.
	if len(cids) > 0 {
		a := api{vars: h.vars, cid: cids[0].String()}
		switch WatchEventType(formsEvent["event_type"].(string)) {
		case WatchSchemaChanges:
			form, err := a.getForm(ctx)
			if err != nil {
				l.Error("Failed to get form", zap.Error(err))
				// Don't abort, dispatch the event without this data.
			}
			formsEvent["form"] = form

		case WatchNewResponses:
			responses, err := a.listResponses(ctx)
			if err != nil {
				l.Error("Failed to list responses", zap.Error(err))
				// Don't abort, dispatch the event without this data.
			}
			formsEvent["response"] = lastResponse(responses)
		}
	}

	// Convert the raw data to an AutoKitteh event.
	wrapped, err := sdktypes.DefaultValueWrapper.Wrap(formsEvent)
	if err != nil {
		l.Error("Failed to wrap Google Forms event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Google Forms event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: strings.ToLower(formsEvent["event_type"].(string)),
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
}

func lastResponse(responses []*forms.FormResponse) *forms.FormResponse {
	if len(responses) == 0 {
		return &forms.FormResponse{}
	}

	last := responses[0]
	for _, r := range responses {
		if r.LastSubmittedTime > last.LastSubmittedTime {
			last = r
		}
	}

	return last
}

func (h handler) dispatchAsyncEventsToConnections(ctx context.Context, name sdktypes.Symbol, watchID string, formsEvent map[string]any) error {
	l := extrazap.ExtractLoggerFromContext(ctx)

	cids, err := h.vars.FindConnectionIDs(ctx, integrationID, name, watchID)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		return err
	}

	akEvent, err := h.constructEvent(ctx, formsEvent, cids)
	if err != nil {
		return err
	}

	for _, cid := range cids {
		eid, err := h.dispatcher.Dispatch(ctx, akEvent.WithConnectionID(cid), nil)
		l := l.With(
			zap.String("connectionID", cid.String()),
			zap.String("eventID", eid.String()),
		)
		if err != nil {
			l.Error("Event dispatch failed", zap.Error(err))
			return err
		}
		l.Debug("Event dispatched")
	}

	return nil
}
