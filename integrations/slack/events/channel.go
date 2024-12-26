package events

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		invalidEventError(l, w, body, err)
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

// Workaround for ENG-980: is_moved should be boolean, but is sometimes 0.
type BoolOrInt bool

// UnmarshalJSON replaces [json.Unmarshal] to support 0 and 1 as false and true.
func (b *BoolOrInt) UnmarshalJSON(data []byte) error {
	v, err := strconv.ParseBool(string(data))
	if err == nil {
		return err
	}
	*b = (BoolOrInt)(v)
	return nil
}

// https://api.slack.com/events/channel_archive
// https://api.slack.com/events/channel_unarchive
// https://api.slack.com/events/group_archive
// https://api.slack.com/events/group_open
// https://api.slack.com/events/group_unarchive
// https://api.slack.com/events/member_joined_channel
type ChannelGroupEvent struct {
	Type        string    `json:"type,omitempty"`
	Channel     string    `json:"channel,omitempty"`
	ChannelType string    `json:"channel_type,omitempty"`
	User        string    `json:"user,omitempty"`
	Inviter     string    `json:"inviter,omitempty"`
	IsMoved     BoolOrInt `json:"is_moved,omitempty"`
	EventTS     string    `json:"event_ts,omitempty"`
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
		invalidEventError(l, w, body, err)
		return nil
	}
	return j.Event
}

// https://api.slack.com/events/member_left_channel
type MemberLeftChannelEvent struct {
	Type        string `json:"type,omitempty"`
	User        string `json:"user,omitempty"`
	Channel     string `json:"channel,omitempty"`
	ChannelType string `json:"channel_type,omitempty"`
	Team        string `json:"team,omitempty"`
	EventTS     string `json:"event_ts,omitempty"`
}

type MemberLeftChannelContainer struct {
	Event *MemberLeftChannelEvent `json:"event"`
}

// https://api.slack.com/events/member_left_channel
func MemberLeftChannelHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	j := &MemberLeftChannelContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}
	return j.Event
}
