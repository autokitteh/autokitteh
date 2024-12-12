package events

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// https://api.slack.com/events/app_mention
type AppMentionEvent struct {
	Type string `json:"type,omitempty"`

	User    string `json:"user,omitempty"`
	Team    string `json:"team,omitempty"`
	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
	EventTS string `json:"event_ts,omitempty"`

	Text   string           `json:"text,omitempty"`
	Blocks []map[string]any `json:"blocks,omitempty"`

	ClientMsgID string `json:"client_msg_id,omitempty"`
}

type appMentionContainer struct {
	Event *AppMentionEvent `json:"event"`
}

// https://api.slack.com/events/app_mention
func AppMentionHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Ignore self-triggered events.
	for _, a := range cb.Authorizations {
		if a.UserID == cb.Event.User {
			l.Debug("Ignoring self-triggered event")
			return nil
		}
	}

	// Parse and return the inner event details.
	j := &appMentionContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}
	return j.Event
}

// https://api.slack.com/events/app_home_opened
// https://api.slack.com/surfaces/app-home
func AppHomeOpenedHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Parse and return the inner event details.
	var eventData map[string]any
	if err := json.Unmarshal(body, &eventData); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}

	event, ok := eventData["event"]; ok
	if !ok {
		l.Warn("Event field not found in the Slack event app_home_opened", zap.ByteString("event", eventData))
		return nil
	}

	return event
}
