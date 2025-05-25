package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/telegram/api"
	"go.autokitteh.dev/autokitteh/integrations/telegram/events"
	"go.autokitteh.dev/autokitteh/integrations/telegram/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// Telegram webhook path
	UpdatePath = "/telegram/webhook"

	// Telegram-specific headers
	headerTelegramSignature = "X-Telegram-Bot-Api-Secret-Token"
)

// handler implements HTTP webhooks to receive and
// dispatch asynchronous event notifications from Telegram.
type handler struct {
	logger        *zap.Logger
	vars          sdkservices.Vars
	dispatch      sdkservices.DispatchFunc
	integrationID sdktypes.IntegrationID
}

func NewHandler(l *zap.Logger, v sdkservices.Vars, d sdkservices.DispatchFunc, i sdktypes.IntegrationID) handler {
	return handler{logger: l, vars: v, dispatch: d, integrationID: i}
}

func (h *handler) dispatchAsyncEventsToConnections(ctx context.Context, cids []sdktypes.ConnectionID, e sdktypes.Event) {
	for _, cid := range cids {
		_, err := h.dispatch(ctx, e.WithConnectionDestinationID(cid), nil)
		if err != nil {
			h.logger.Warn("event dispatch failed", zap.Error(err), zap.String("connection_id", cid.String()))
		}
	}
}

// checkRequest checks that the given HTTP request has a valid content type and
// optionally a valid webhook secret signature, and if so it returns the request's body.
func (h handler) checkRequest(w http.ResponseWriter, r *http.Request, l *zap.Logger) []byte {
	// "Content-Type" header.
	gotContentType := r.Header.Get(common.HeaderContentType)
	wantContentType := "application/json"
	if gotContentType == "" || gotContentType != wantContentType {
		l.Warn("incoming event: unexpected header value",
			zap.String("header", common.HeaderContentType),
			zap.String("got", gotContentType),
			zap.String("want", wantContentType),
		)
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	// Read the request body.
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Warn("incoming event: failed to read request body", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return nil
	}

	// Optional webhook secret verification
	if secret := r.Header.Get(headerTelegramSignature); secret != "" {
		// TODO: Implement webhook secret verification when configured
		l.Debug("webhook secret header present", zap.String("secret", secret))
	}

	return body
}

// findConnectionID finds the connection ID for the incoming Telegram update.
// It uses the bot token to identify the connection.
func (h handler) findConnectionID(ctx context.Context, update events.Update, l *zap.Logger) (sdktypes.ConnectionID, error) {
	// We don't have the bot token in the update, so we need to match based on bot info
	// For now, let's find all connections and match by bot username or ID if available
	cids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, vars.BotTokenVar, "")
	if err != nil {
		return sdktypes.InvalidConnectionID, fmt.Errorf("failed to find connection IDs: %w", err)
	}

	if len(cids) == 0 {
		return sdktypes.InvalidConnectionID, fmt.Errorf("no Telegram connections found")
	}

	// For now, return the first connection. In a real implementation,
	// we would need to store bot info to match the specific bot.
	return cids[0], nil
}

// Helper to wrap a Telegram message as an event (since events.WrapMessage does not exist)
func wrapMessage(msg *api.Message, cid sdktypes.ConnectionID) (sdktypes.Event, error) {
	// Convert api.Message to events.Message via JSON marshal/unmarshal
	b, err := json.Marshal(msg)
	if err != nil {
		return sdktypes.InvalidEvent, err
	}
	var em events.Message
	if err := json.Unmarshal(b, &em); err != nil {
		return sdktypes.InvalidEvent, err
	}
	wrapped, err := sdktypes.WrapValue(em)
	if err != nil {
		return sdktypes.InvalidEvent, err
	}
	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		return sdktypes.InvalidEvent, err
	}
	return sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: "telegram_message",
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
}

// func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
func (h *handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("url_path", UpdatePath))

	// Read the request body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		l.Error("Failed to read request body", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	var update api.Update
	if err := json.Unmarshal(body, &update); err != nil {
		l.Error("Failed to parse webhook payload", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	var akEvent sdktypes.Event
	var eventType string

	switch {
	case update.Message != nil:
		eventType = "message"
		akEvent, err = h.transformMessage(update.Message)
	case update.CallbackQuery != nil:
		eventType = "callback_query"
		akEvent, err = h.transformCallbackQuery(update.CallbackQuery)
	case update.EditedMessage != nil:
		eventType = "edited_message"
		akEvent, err = h.transformMessage(update.EditedMessage)
	default:
		l.Debug("Ignoring unsupported update type")
		w.WriteHeader(http.StatusOK)
		return
	}

	if err != nil {
		l.Error("Failed to transform update", zap.String("type", eventType), zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Find connections and dispatch event
	cids, err := h.findConnectionsForBot(r.Context())
	if err != nil {
		l.Error("Failed to find connections", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	h.dispatchAsyncEventsToConnections(r.Context(), cids, akEvent)

	// Handle callback query acknowledgment
	if update.CallbackQuery != nil {
		h.answerCallbackQuery(r.Context(), update.CallbackQuery)
	}

	w.WriteHeader(http.StatusOK)
}

// transformMessage transforms a Telegram message into an AutoKitteh event.
func (h *handler) transformMessage(msg *api.Message) (sdktypes.Event, error) {
	cids, err := h.findConnectionsForBot(context.Background())
	if err != nil || len(cids) == 0 {
		return sdktypes.InvalidEvent, fmt.Errorf("no Telegram connections found: %w", err)
	}
	return wrapMessage(msg, cids[0])
}

// transformCallbackQuery transforms a Telegram callback query into an AutoKitteh event.
func (h *handler) transformCallbackQuery(callback *api.CallbackQuery) (sdktypes.Event, error) {
	wrapped, err := sdktypes.WrapValue(callback)
	if err != nil {
		return sdktypes.InvalidEvent, err
	}
	data, err := wrapped.ToStringValuesMap()
	if err != nil {
		return sdktypes.InvalidEvent, err
	}
	return sdktypes.EventFromProto(&sdktypes.EventPB{
		EventType: "callback_query",
		Data:      kittehs.TransformMapValues(data, sdktypes.ToProto),
	})
}

// findConnectionsForBot finds all connection IDs for the current integration.
func (h *handler) findConnectionsForBot(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	cids, err := h.vars.FindConnectionIDs(ctx, h.integrationID, vars.BotTokenVar, "")
	if err != nil {
		return nil, fmt.Errorf("failed to find connection IDs: %w", err)
	}
	if len(cids) == 0 {
		return nil, fmt.Errorf("no Telegram connections found")
	}
	return cids, nil
}

// answerCallbackQuery acknowledges a callback query using the bot token from the first connection.
func (h *handler) answerCallbackQuery(ctx context.Context, callback *api.CallbackQuery) {
	cids, err := h.findConnectionsForBot(ctx)
	if err != nil || len(cids) == 0 {
		h.logger.Warn("No Telegram connections found for answering callback query", zap.Error(err))
		return
	}
	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cids[0]))
	if err != nil {
		h.logger.Warn("Failed to get vars for connection", zap.Error(err))
		return
	}
	token := vs.GetValue(vars.BotTokenVar)
	if token == "" {
		h.logger.Warn("Bot token not found in connection vars")
		return
	}
	client := api.NewClient(token)
	err = client.AnswerCallbackQuery(ctx, callback.ID, "Button pressed!", false)
	if err != nil {
		h.logger.Warn("Failed to answer callback query", zap.Error(err))
	}
}
