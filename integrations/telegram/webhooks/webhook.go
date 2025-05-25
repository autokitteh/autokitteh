package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/telegram/events"
	"go.autokitteh.dev/autokitteh/integrations/telegram/vars"
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

// HandleUpdate handles incoming Telegram updates (webhooks).
func (h handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("handler", "update"))

	// Check and read the request.
	body := h.checkRequest(w, r, l)
	if body == nil {
		return // checkRequest already sent an HTTP error response.
	}

	// Parse the Telegram update.
	var update events.Update
	if err := json.Unmarshal(body, &update); err != nil {
		l.Warn("incoming event: invalid JSON", zap.Error(err))
		common.HTTPError(w, http.StatusBadRequest)
		return
	}

	l = l.With(zap.Int64("update_id", update.UpdateID))

	// Find the relevant AutoKitteh connection.
	ctx := r.Context()
	cid, err := h.findConnectionID(ctx, update, l)
	if err != nil {
		l.Warn("incoming event: connection not found", zap.Error(err))
		common.HTTPError(w, http.StatusNotFound)
		return
	}

	l = l.With(zap.String("connection_id", cid.String()))

	// Dispatch the event to AutoKitteh.
	wrapped, err := events.WrapUpdate(update, cid, h.integrationID)
	if err != nil {
		l.Error("incoming event: failed to wrap", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	eid, err := h.dispatch(ctx, wrapped, nil)
	if err != nil {
		l.Error("incoming event: dispatch failed", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	l.Info("incoming event dispatched", zap.String("event_id", eid.String()))

	// Telegram expects a 200 OK response to acknowledge receipt.
	w.WriteHeader(http.StatusOK)
}
