package telegram

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// HandleWebhook processes incoming Telegram webhook events
func (h handler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("method", "HandleWebhook"))

	// Validate content type
	if common.PostWithoutJSONContentType(r) {
		l.Warn("Unexpected content type", zap.String("content_type", r.Header.Get("Content-Type")))
		http.Error(w, "Expected JSON content type", http.StatusBadRequest)
		return
	}

	// Parse the webhook payload
	var update TelegramUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		l.Warn("Failed to decode webhook payload", zap.Error(err))
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Dispatch the event based on type
	if err := h.dispatchEvent(r.Context(), &update); err != nil {
		l.Error("Failed to dispatch Telegram event", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Acknowledge the webhook
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// dispatchEvent dispatches different types of Telegram events
func (h handler) dispatchEvent(ctx context.Context, update *TelegramUpdate) error {
	var eventType string
	var data map[string]any

	switch {
	case update.Message != nil:
		eventType = "message"
		data = h.messageToData(update.Message, update.UpdateID)
	case update.Edited != nil:
		eventType = "edited_message"
		data = h.messageToData(update.Edited, update.UpdateID)
	case update.Callback != nil:
		eventType = "callback_query"
		data = h.callbackToData(update.Callback, update.UpdateID)
	default:
		h.logger.Debug("Unhandled update type", zap.Int("update_id", update.UpdateID))
		return nil
	}

	// Transform the Telegram event into an AutoKitteh event
	akEvent, err := common.TransformEvent(h.logger, data, eventType)
	if err != nil {
		return err
	}

	// Find connections for this integration
	connections, err := h.findConnections(ctx)
	if err != nil {
		return err
	}

	// Dispatch to all connections
	common.DispatchEvent(ctx, h.logger, h.dispatch, akEvent, connections)

	return nil
}

// messageToData converts a Telegram message to event data
func (h handler) messageToData(msg *TelegramMessage, updateID int) map[string]any {
	data := map[string]any{
		"update_id":  updateID,
		"message_id": msg.MessageID,
		"date":       msg.Date,
		"chat": map[string]any{
			"id":   msg.Chat.ID,
			"type": msg.Chat.Type,
		},
	}

	if msg.Chat.Title != "" {
		data["chat"].(map[string]any)["title"] = msg.Chat.Title
	}
	if msg.Chat.Username != "" {
		data["chat"].(map[string]any)["username"] = msg.Chat.Username
	}

	if msg.From != nil {
		data["from"] = map[string]any{
			"id":         msg.From.ID,
			"is_bot":     msg.From.IsBot,
			"first_name": msg.From.FirstName,
		}
		if msg.From.Username != "" {
			data["from"].(map[string]any)["username"] = msg.From.Username
		}
	}

	if msg.Text != "" {
		data["text"] = msg.Text
	}

	if len(msg.Photo) > 0 {
		photos := make([]map[string]any, len(msg.Photo))
		for i, photo := range msg.Photo {
			photos[i] = map[string]any{
				"file_id": photo.FileID,
				"width":   photo.Width,
				"height":  photo.Height,
			}
			if photo.FileSize > 0 {
				photos[i]["file_size"] = photo.FileSize
			}
		}
		data["photo"] = photos
	}

	if msg.Document != nil {
		doc := map[string]any{
			"file_id": msg.Document.FileID,
		}
		if msg.Document.FileName != "" {
			doc["file_name"] = msg.Document.FileName
		}
		if msg.Document.MimeType != "" {
			doc["mime_type"] = msg.Document.MimeType
		}
		if msg.Document.FileSize > 0 {
			doc["file_size"] = msg.Document.FileSize
		}
		data["document"] = doc
	}

	return data
}

// callbackToData converts a Telegram callback query to event data
func (h handler) callbackToData(cb *TelegramCallback, updateID int) map[string]any {
	data := map[string]any{
		"update_id": updateID,
		"id":        cb.ID,
		"from": map[string]any{
			"id":         cb.From.ID,
			"is_bot":     cb.From.IsBot,
			"first_name": cb.From.FirstName,
		},
	}

	if cb.From.Username != "" {
		data["from"].(map[string]any)["username"] = cb.From.Username
	}

	if cb.Data != "" {
		data["data"] = cb.Data
	}

	if cb.Message != nil {
		data["message"] = h.messageToData(cb.Message, 0)
	}

	return data
}

// findConnections finds all connections for this integration
func (h handler) findConnections(ctx context.Context) ([]sdktypes.ConnectionID, error) {
	// Find all connections for this integration
	return h.vars.FindConnectionIDs(ctx, integrationID, BotToken, "")
}
