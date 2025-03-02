// Package chat is a lightweight wrapper over the "chat" methods
// in Slack's Web API: https://api.slack.com/methods?filter=chat.
package chat

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/chat.update#examples
type UpdateResponse struct {
	api.SlackResponse

	Channel string   `json:"channel,omitempty"`
	TS      string   `json:"ts,omitempty"`
	Text    string   `json:"text,omitempty"`
	Message *Message `json:"message,omitempty"`
}

// -------------------- Auxiliary data structures --------------------

type BotProfile struct {
	ID     string `json:"id,omitempty"`
	AppID  string `json:"app_id,omitempty"`
	TeamID string `json:"team_id,omitempty"`

	Name string `json:"name,omitempty"`

	Deleted bool `json:"deleted,omitempty"`
	Updated int  `json:"updated,omitempty"`
}

type Edited struct {
	User string `json:"user,omitempty"`
	TS   string `json:"ts,omitempty"`
}

// https://api.slack.com/types/conversation
// https://api.slack.com/methods/conversations.history#examples
// https://api.slack.com/methods/conversations.replies#examples
// https://github.com/slackapi/slack-api-specs/blob/master/web-api ("objs_message")
// https://api.slack.com/events/message
type Message struct {
	Type    string `json:"type,omitempty"`
	Subtype string `json:"subtype,omitempty"`
	// https://api.slack.com/events/message#hidden_subtypes
	Hidden bool `json:"hidden,omitempty"`

	Text   string           `json:"text,omitempty"`
	Blocks []map[string]any `json:"blocks,omitempty"`
	Edited *Edited          `json:"edited,omitempty"`

	User         string      `json:"user,omitempty"`
	AppID        string      `json:"app_id,omitempty"`
	BotID        string      `json:"bot_id,omitempty"`
	BotProfile   *BotProfile `json:"bot_profile,omitempty"`
	ParentUserID string      `json:"parent_user_id,omitempty"`

	Team        string `json:"team,omitempty"`
	Channel     string `json:"channel,omitempty"`
	ChannelType string `json:"channel_type,omitempty"`
	TS          string `json:"ts,omitempty"`
	EventTS     string `json:"event_ts,omitempty"`
	Permalink   string `json:"permalink,omitempty"`

	// https://api.slack.com/types/conversation
	// https://api.slack.com/methods/conversations.replies#examples
	ReplyCount      int      `json:"reply_count,omitempty"`
	ReplyUsersCount int      `json:"reply_users_count,omitempty"`
	LatestReply     string   `json:"latest_reply,omitempty"`
	ReplyUsers      []string `json:"reply_users,omitempty"`
	LastRead        string   `json:"last_read,omitempty"`
	UnreadCount     int      `json:"unread_count,omitempty"`
	// A count of messages that the calling user has yet to read that
	// matter to them (excludes things like join/leave messages).
	UnreadCountDisplay int  `json:"unread_count_display,omitempty"`
	IsLocked           bool `json:"is_locked,omitempty"`
	Subscribed         bool `json:"subscribed,omitempty"`

	// https://api.slack.com/events/message#stars
	IsStarred bool       `json:"is_starred,omitempty"`
	PinnedTo  []string   `json:"pinned_to,omitempty"`
	Reactions []Reaction `json:"reactions,omitempty"`

	// https://api.slack.com/events/message/channel_join
	Inviter string `json:"inviter,omitempty"`
	// https://api.slack.com/events/message/channel_name
	Name    string `json:"name,omitempty"`
	OldName string `json:"old_name,omitempty"`
	// https://api.slack.com/events/message/channel_purpose
	Purpose string `json:"purpose,omitempty"`
	// https://api.slack.com/events/message/channel_topic
	Topic string `json:"topic,omitempty"`

	// TODO: https://api.slack.com/events/message/file_share

	// https://api.slack.com/events/message/message_changed
	Message         *Message `json:"message,omitempty"`
	PreviousMessage *Message `json:"previous_message,omitempty"`
	// https://api.slack.com/events/message/message_deleted
	DeletedTS string `json:"deleted_ts,omitempty"`
	// https://api.slack.com/events/message/message_replied
	ThreadTS string `json:"thread_ts,omitempty"`

	// https://api.slack.com/events/message/thread_broadcast
	Root *Message `json:"root,omitempty"`

	ClientMsgID string `json:"client_msg_id,omitempty"`
}

// https://api.slack.com/events/message
type Reaction struct {
	Name  string   `json:"name,omitempty"`
	Users []string `json:"users,omitempty"`
	Count int      `json:"count,omitempty"`
}

// https://api.slack.com/reference/block-kit/composition-objects#text
type Text struct {
	// Type must be "mrkdwn" or "plain_text".
	Type string `json:"type,omitempty"`
	// Text from a minimum of 1 to a maximum of 3000 characters.
	// See https://api.slack.com/reference/surfaces/formatting.
	Text string `json:"text,omitempty"`
	// This field is only usable when [Type] is "plain_text".
	Emoji bool `json:"emoji,omitempty"`
	// When set to false (as is default) URLs will be auto-converted into
	// links, conversation names will be link-ified, and certain mentions will
	// be automatically parsed. Using a value of true will skip any preprocessing
	// of this nature, although you can still include manual parsing strings.
	// This field is only usable when Type is "mrkdwn". See
	// https://api.slack.com/reference/surfaces/formatting.
	Verbatim bool `json:"verbatim,omitempty"`
}
