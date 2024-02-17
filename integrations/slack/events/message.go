package events

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/api/chat"
)

type messageContainer struct {
	Event *chat.Message `json:"event"`
}

// https://api.slack.com/events/message
// https://api.slack.com/events/message.channels
// https://api.slack.com/events/message.groups
// https://api.slack.com/events/message.im
// https://api.slack.com/events/message.mpim
func MessageHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Ignore self-triggered events.
	for _, a := range cb.Authorizations {
		if a.UserID == cb.Event.User {
			l.Debug("Ignoring self-triggered event")
			return nil
		}
	}

	// Parse the inner event details.
	j := &messageContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		l.Error("Failed to parse JSON payload",
			zap.Error(err),
			zap.ByteString("json", body),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}

	// Also ignore self-triggered "message_changed" events.
	if j.Event.User == "" && j.Event.Message != nil {
		for _, a := range cb.Authorizations {
			if a.UserID == j.Event.Message.User {
				l.Debug(`Ignoring self-triggered "message_changed" event`)
				return nil
			}
		}
	}

	return j.Event
}
