package events

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

// https://api.slack.com/events/reaction_added
// https://api.slack.com/events/reaction_removed
type ReactionEvent struct {
	Type string `json:"type,omitempty"`
	User string `json:"user,omitempty"`
	// Emoji name (without ":" on either side).
	Reaction string `json:"reaction,omitempty"`
	// Item is a brief reference to what was reacted to.
	Item *Item `json:"item,omitempty"`
	// ItemUser is the ID of the user that created the original item that has
	// been reacted to. Some messages aren't authored by "users," like those
	// created by incoming webhooks (https://api.slack.com/messaging/webhooks).
	// Events related to these messages will not include an [ItemUser].
	ItemUser string `json:"item_user,omitempty"`
	EventTS  string `json:"event_ts,omitempty"`
}

// Item is a brief reference to what was reacted to.
type Item struct {
	Type        string `json:"type,omitempty"`
	Channel     string `json:"channel,omitempty"`
	TS          string `json:"ts,omitempty"`
	File        string `json:"file,omitempty"`
	FileComment string `json:"file_comment,omitempty"`
}

type reactionContainer struct {
	Event *ReactionEvent `json:"event"`
}

// https://api.slack.com/events/reaction_added
// https://api.slack.com/events/reaction_removed
func ReactionHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Ignore self-triggered events.
	for _, a := range cb.Authorizations {
		if a.UserID == cb.Event.User {
			l.Debug("Ignoring self-triggered event")
			return nil
		}
	}

	// Parse and return the inner event details.
	j := &reactionContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}
	return j.Event
}
