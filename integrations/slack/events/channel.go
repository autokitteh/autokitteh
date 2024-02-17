package events

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/api/conversations"
)

// https://api.slack.com/events/channel_created
type ChannelCreatedEvent struct {
	Type    string                 `json:"type,omitempty"`
	Channel *conversations.Channel `json:"channel,omitempty"`
	EventTS string                 `json:"event_ts,omitempty"`
}

type channelCreatedContainer struct {
	Event *ChannelCreatedEvent `json:"event"`
}

// https://api.slack.com/events/channel_created
func ChannelCreatedHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Parse the inner event details.
	j := &channelCreatedContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		l.Error("Failed to parse JSON payload",
			zap.Error(err),
			zap.ByteString("json", body),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}

	// Ignore self-triggered events.
	for _, a := range cb.Authorizations {
		if a.UserID == j.Event.Channel.Creator {
			l.Debug("Ignoring self-triggered event")
			return nil
		}
	}

	// Return the inner event details.
	return j.Event
}

// https://api.slack.com/events/channel_archive
// https://api.slack.com/events/channel_unarchive
// https://api.slack.com/events/group_archive
// https://api.slack.com/events/group_open
// https://api.slack.com/events/group_unarchive
// https://api.slack.com/events/member_joined_channel
type ChannelGroupEvent struct {
	Type        string `json:"type,omitempty"`
	Channel     string `json:"channel,omitempty"`
	ChannelType string `json:"channel_type,omitempty"`
	User        string `json:"user,omitempty"`
	Inviter     string `json:"inviter,omitempty"`
	IsMoved     bool   `json:"is_moved,omitempty"`
	EventTS     string `json:"event_ts,omitempty"`
}

type channelGroupContainer struct {
	Event *ChannelGroupEvent `json:"event"`
}

// https://api.slack.com/events/channel_archive
// https://api.slack.com/events/channel_unarchive
// https://api.slack.com/events/group_archive
// https://api.slack.com/events/group_open
// https://api.slack.com/events/group_unarchive
// https://api.slack.com/events/member_joined_channel
func ChannelGroupHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Ignore self-triggered events.
	for _, a := range cb.Authorizations {
		if cb.Event.Inviter != "" && a.UserID == cb.Event.Inviter {
			l.Debug("Ignoring self-triggered event")
			return nil
		}
		if cb.Event.Inviter == "" && a.UserID == cb.Event.User {
			l.Debug("Ignoring self-triggered event")
			return nil
		}
	}

	// Parse and return the inner event details.
	j := &channelGroupContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		l.Error("Failed to parse JSON payload",
			zap.Error(err),
			zap.ByteString("json", body),
		)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return nil
	}
	return j.Event
}
