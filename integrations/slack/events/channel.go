package events

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/slack/api/conversations"
)

// https://api.slack.com/events/channel_created
// https://api.slack.com/events/channel_rename
// https://api.slack.com/events/group_rename
type ChannelCreatedRenameEvent struct {
	Type    string                 `json:"type,omitempty"`
	Channel *conversations.Channel `json:"channel,omitempty"`
	EventTS string                 `json:"event_ts,omitempty"`
}

type channelCreatedContainer struct {
	Event *ChannelCreatedRenameEvent `json:"event"`
}

// https://api.slack.com/events/channel_created
// https://api.slack.com/events/channel_rename
// https://api.slack.com/events/group_rename
func ChannelCreatedRenameHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
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
// https://api.slack.com/events/member_left_channel
type ChannelGroupMemberEvent struct {
	Type        string    `json:"type,omitempty"`
	Channel     string    `json:"channel,omitempty"`
	ChannelType string    `json:"channel_type,omitempty"`
	User        string    `json:"user,omitempty"`
	Inviter     string    `json:"inviter,omitempty"`
	IsMoved     BoolOrInt `json:"is_moved,omitempty"`
	EventTS     string    `json:"event_ts,omitempty"`
	Team        string    `json:"team,omitempty"`
	Enterprise  string    `json:"enterprise,omitempty"`
}

type channelGroupMemberContainer struct {
	Event *ChannelGroupMemberEvent `json:"event"`
}

// https://api.slack.com/events/channel_archive
// https://api.slack.com/events/channel_unarchive
// https://api.slack.com/events/group_archive
// https://api.slack.com/events/group_open
// https://api.slack.com/events/group_unarchive
// https://api.slack.com/events/member_joined_channel
// https://api.slack.com/events/member_left_channel
func ChannelGroupMemberHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
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
	j := &channelGroupMemberContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}
	return j.Event
}

// https://api.slack.com/events/tokens_revoked
type TokensRevokedEvent struct {
	Type    string `json:"type,omitempty"`
	Tokens  Tokens `json:"tokens,omitempty"`
	EventTS string `json:"event_ts,omitempty"`
}

type Tokens struct {
	OAuth []string `json:"oauth,omitempty"`
	Bot   []string `json:"bot,omitempty"`
}

type tokensRevokedContainer struct {
	Event *TokensRevokedEvent `json:"event"`
}

// https://api.slack.com/events/tokens_revoked
func TokensRevokedHandler(l *zap.Logger, w http.ResponseWriter, body []byte, cb *Callback) any {
	// Parse and return the inner event details.
	j := &tokensRevokedContainer{}
	if err := json.Unmarshal(body, j); err != nil {
		invalidEventError(l, w, body, err)
		return nil
	}
	return j.Event
}
