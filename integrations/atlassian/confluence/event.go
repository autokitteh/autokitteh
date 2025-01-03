package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/internal/extrazap"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	headerContentType   = "Content-Type"
	headerAuthorization = "Authorization"
	contentTypeJSON     = "application/json"
)

// Default HTTP client with a timeout for short-lived HTTP requests.
var httpClient = http.Client{Timeout: 3 * time.Second}

// handler is an autokitteh webhook which implements [http.Handler]
// to receive and dispatch asynchronous event notifications.
type handler struct {
	logger   *zap.Logger
	oauth    sdkservices.OAuth
	vars     sdkservices.Vars
	dispatch sdkservices.DispatchFunc
}

func NewHTTPHandler(l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars, d sdkservices.DispatchFunc) handler {
	return handler{logger: l, oauth: o, vars: v, dispatch: d}
}

// handleEvent receives from Atlassian asynchronous events,
// and dispatches them to zero or more AutoKitteh connections.
// Note 1: By default, AutoKitteh creates webhooks automatically,
// subscribing to all events - see "webhooks.go" for more details.
// Note 2: Unlike Jira webhooks, auto-created Confluence webhooks
// do not expire after 30 days, so no need to extend their deadline.
// Note 3: The requests are sent by a service, so no need to respond
// with user-friendly error web pages.
func (h handler) handleEvent(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// TODO(ENG-1081): Verify the HMAC signature in "X-Hub-Signature"
	// (Confluence doesn't recognize secrets when creating webhooks).

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeJSON) {
		l.Warn("Incoming Atlassian event with bad header", zap.String(headerContentType, contentType))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Parse some of the metadata in the Atlassian event's JSON content.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Warn("Failed to read content of incoming Atlassian event", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var atlassianEvent map[string]any
	if err := json.Unmarshal(body, &atlassianEvent); err != nil {
		l.Warn("Failed to unmarshal JSON in incoming Atlassian event", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Determine the event type based on the webhook callback URL, and the content
	// of the event (unlike Jira, Confluence events don't have an event type field).
	eventType, ok := extractEntityType(l, atlassianEvent, r.PathValue("category"))
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Iterate through all the relevant connections for this event.
	atlassianURL := extractBaseURL(atlassianEvent)

	// Construct an AutoKitteh event from the Atlassian event.
	akEvent, err := constructEvent(l, atlassianEvent, eventType)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	ctx := extrazap.AttachLoggerToContext(l, r.Context())
	cids, err := h.vars.FindConnectionIDs(ctx, integrationID, baseURL, atlassianURL)
	if err != nil {
		l.Error("Failed to find connection IDs", zap.Error(err))
		return
	}

	// Dispatch the event to all of them, for asynchronous handling.
	h.dispatchAsyncEventsToConnections(ctx, cids, akEvent)

	// Returning immediately without an error = acknowledgement of receipt.
}

// All non-"content" event types (with the created/updated/removed suffix).
// "content" is actually included, to support "content_updated" events.
// https://developer.atlassian.com/cloud/confluence/modules/webhook/
// https://confluence.atlassian.com/conf715/managing-webhooks-1096098349.html
var simpleEntities = []string{
	"attachment", "blog", "blueprint_page", "comment",
	"content", "group", "label", "page", "relation", "space",
}

func extractEntityType(l *zap.Logger, atlassianEvent map[string]any, category string) (string, bool) {
	if category == "" {
		l.Warn("Unexpected Confluence event callback: missing category in URL")
		return "", false
	}

	for _, entity := range simpleEntities {
		if _, ok := atlassianEvent[entity]; ok {
			return fmt.Sprintf("%s_%s", entity, category), true
		}
	}

	// Some "attachment_*" and "relation_*" events look a little different.
	if _, ok := atlassianEvent["attachments"]; ok {
		return "attachment_" + category, true
	}
	if _, ok := atlassianEvent["relationData"]; ok {
		return "relation_" + category, true
	}

	// Last but not least: "content_*" events.
	if contentType, ok := atlassianEvent["type"]; ok {
		if s, ok := contentType.(string); ok && strings.Contains(s, "content") {
			return "content_" + category, true
		}
	}

	l.Error("Unrecognized Confluence event",
		zap.String("category", category),
		zap.Any("event", atlassianEvent),
	)
	return "", false
}

func extractBaseURL(atlassianEvent map[string]any) string {
	for _, entity := range simpleEntities {
		if data, ok := atlassianEvent[entity].(map[string]any); ok {
			if url, ok := data["self"].(string); ok {
				return strings.Split(url, "/wiki")[0]
			}
		}
	}

	// Last but not least: "content_*" and "relation_*" events.
	if links, ok := atlassianEvent["_links"].(map[string]string); ok {
		if base, ok := links["base"]; ok {
			return strings.Split(base, "/wiki")[0]
		}
	}

	return ""
}

func constructEvent(l *zap.Logger, atlassianEvent map[string]any, eventType string) (sdktypes.Event, error) {
	l = l.With(zap.Any("event", atlassianEvent))

	wrapped, err := sdktypes.WrapValue(atlassianEvent)
	if err != nil {
		l.Error("Failed to wrap Atlassian event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		l.Error("Failed to convert wrapped Atlassian event", zap.Error(err))
		return sdktypes.InvalidEvent, err
	}

	akEvent, err := sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: eventType,
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
	if err != nil {
		l.Error("Failed to convert protocol buffer to SDK event",
			zap.String("eventType", eventType),
			zap.Any("data", data),
			zap.Error(err),
		)
		return sdktypes.InvalidEvent, err
	}

	return akEvent, nil
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
