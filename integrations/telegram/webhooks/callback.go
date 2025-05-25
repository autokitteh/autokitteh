package webhooks

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/telegram/api"
)

const CallbackPath = "/telegram/callback"

// HandleCallbackQuery handles inline keyboard button presses (callback queries)
func (h *handler) HandleCallbackQuery(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("url_path", CallbackPath))

	// Parse the webhook payload
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

	// Only handle callback queries
	if update.CallbackQuery == nil {
		l.Debug("Ignoring non-callback update")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Transform the callback query into an AutoKitteh event
	akEvent, err := h.transformCallbackQuery(update.CallbackQuery)
	if err != nil {
		l.Error("Failed to transform callback query", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Find connections for this bot
	cids, err := h.findConnectionsForBot(r.Context())
	if err != nil {
		l.Error("Failed to find connections", zap.Error(err))
		common.HTTPError(w, http.StatusInternalServerError)
		return
	}

	// Dispatch the event
	h.dispatchAsyncEventsToConnections(r.Context(), cids, akEvent)

	// Answer the callback query to acknowledge it
	h.answerCallbackQuery(r.Context(), update.CallbackQuery)

	w.WriteHeader(http.StatusOK)
}
