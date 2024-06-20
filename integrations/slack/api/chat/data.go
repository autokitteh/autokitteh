// Package chat is a lightweight wrapper over the "chat" methods
// in Slack's Web API: https://api.slack.com/methods?filter=chat.
package chat

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/chat.delete#args
type DeleteRequest struct {
	// Channel containing the message to be deleted. Required.
	// https://api.slack.com/methods/chat.postMessage#channels
	Channel string `json:"channel"`
	// TS is a timestamp of the message to be deleted. Required.
	TS string `json:"ts"`
}

// https://api.slack.com/methods/chat.delete#examples
type DeleteResponse struct {
	api.SlackResponse

	Channel string `json:"channel,omitempty"`
	TS      string `json:"ts,omitempty"`
}

// https://api.slack.com/methods/chat.getPermalink#args
type GetPermalinkResponse struct {
	api.SlackResponse

	Permalink string `json:"permalink,omitempty"`
	Channel   string `json:"channel,omitempty"`
}

// https://api.slack.com/methods/chat.postEphemeral#args
type PostEphemeralRequest struct {
	// https://api.slack.com/methods/chat.postEphemeral#target-channels-and-users
	Channel string `json:"channel"`
	// User ID of the user who will receive the ephemeral message. The user
	// should be in the channel specified by the channel argument. Required.
	User string `json:"user"`

	// https://api.slack.com/reference/surfaces/formatting
	// https://api.slack.com/methods/chat.postEphemeral#markdown
	Text string `json:"text,omitempty"`
	// https://api.slack.com/reference/block-kit
	// https://api.slack.com/messaging/composing/layouts
	// https://api.slack.com/messaging/interactivity
	// https://app.slack.com/block-kit-builder/
	Blocks []map[string]interface{} `json:"blocks,omitempty"`

	// ThreadTS provides another message's [TS] value to make this message
	// a reply. Avoid using a reply's [TS] value; use its parent instead.
	// See https://api.slack.com/methods/chat.postMessage#threads.
	ThreadTS string `json:"thread_ts,omitempty"`
}

// https://api.slack.com/methods/chat.postEphemeral#examples
type PostEphemeralResponse struct {
	api.SlackResponse

	Channel   string `json:"channel,omitempty"`
	MessageTS string `json:"message_ts,omitempty"`
}

// https://api.slack.com/methods/chat.postMessage#args
type PostMessageRequest struct {
	// See https://api.slack.com/methods/chat.postMessage#channels.
	// Slack user ID ("U"), user DM ID ("D"), multi-person/group DM ID ("G"),
	// channel ID ("C"), channel name (with or without the "#" prefix). Note
	// that all targets except "U", "D" and public channels require our Slack
	// app to be added in advance. Required.
	Channel string `json:"channel"`

	// https://api.slack.com/reference/surfaces/formatting
	// https://api.slack.com/methods/chat.postMessage#blocks_and_attachments
	Text string `json:"text,omitempty"`
	// https://api.slack.com/reference/block-kit
	// https://api.slack.com/messaging/composing/layouts
	// https://api.slack.com/messaging/interactivity
	// https://app.slack.com/block-kit-builder/
	Blocks []map[string]interface{} `json:"blocks,omitempty"`

	// ThreadTS provides another message's [TS] value to make this message
	// a reply. Avoid using a reply's [TS] value; use its parent instead.
	// See https://api.slack.com/methods/chat.postMessage#threads.
	ThreadTS string `json:"thread_ts,omitempty"`
	// ReplyBroadcast is used in conjunction with [ThreadTS] and indicates
	// whether the reply should be made visible to everyone in the channel
	// or conversation. Default = false.
	ReplyBroadcast bool `json:"reply_broadcast,omitempty"`

	// Name to display alongside the message, instead of the bot's name.
	Username string `json:"username,omitempty"`
	// URL to an image to use as the user's icon for this message, instead of the bot's.
	IconURL string `json:"icon_url,omitempty"`
}

// https://api.slack.com/methods/chat.postMessage#examples
type PostMessageResponse struct {
	api.SlackResponse

	Channel string   `json:"channel,omitempty"`
	TS      string   `json:"ts,omitempty"`
	Message *Message `json:"message,omitempty"`
}

// https://api.slack.com/methods/chat.update#args
type UpdateRequest struct {
	// See https://api.slack.com/methods/chat.postMessage#channels.
	// Slack user ID ("U"), user DM ID ("D"), multi-person/group DM ID ("G"),
	// channel ID ("C"), channel name (with or without the "#" prefix). Note
	// that all targets except "U", "D" and public channels require our Slack
	// app to be added in advance. Required
	Channel string `json:"channel"`
	// Timestamp of the message to be updated. Required.
	TS string `json:"ts"`

	// https://api.slack.com/reference/surfaces/formatting
	// https://api.slack.com/methods/chat.postMessage#blocks_and_attachments
	Text string `json:"text,omitempty"`
	// https://api.slack.com/reference/block-kit
	// https://api.slack.com/messaging/composing/layouts
	// https://api.slack.com/messaging/interactivity
	// https://app.slack.com/block-kit-builder/
	Blocks []map[string]interface{} `json:"blocks,omitempty"`

	// ReplyBroadcast is used in conjunction with [TS] and indicates whether
	// the reply should be made visible to everyone in the channel or
	// conversation. Defaults to false.
	ReplyBroadcast bool `json:"reply_broadcast,omitempty"`
}

// https://api.slack.com/methods/chat.update#examples
type UpdateResponse struct {
	api.SlackResponse

	Channel string   `json:"channel,omitempty"`
	TS      string   `json:"ts,omitempty"`
	Text    string   `json:"text,omitempty"`
	Message *Message `json:"message,omitempty"`
}

type SendApprovalMessageRequest struct {
	// Slack user ID ("U"), user DM ID ("D"), multi-person/group DM ID ("G"),
	// channel ID ("C"), channel name (with or without the "#" prefix). All
	// targets except "U", "D" and public channels require our Slack app to
	// be added in advance.
	Target string `json:"target"`

	// Header text: up to 150 characters, no markdown, but may include emoji.
	Header string `json:"header,omitempty"`
	// Message text: up to 3000 characters, may include markdown and emoji.
	// See https://api.slack.com/reference/surfaces/formatting.
	Message string `json:"message,omitempty"`
	// Green button text (e.g. "Yes", "Approve", "Proceed"). No markdown, may
	// include emoji. May truncate with ~30 characters, maximum length is 75
	// characters. Default = "Approve".
	GreenButton string `json:"green_button,omitempty"`
	// Red button text (e.g. "No", "Deny", "Reject", "Abort"). No markdown, may
	// include emoji. May truncate with ~30 characters, maximum length is 75
	// characters. Default = "Deny".
	RedButton string `json:"red_button,omitempty"`

	// ThreadTS provides another message's [TS] value to make this message
	// a reply. Avoid using a reply's [TS] value; use its parent instead.
	// See https://api.slack.com/methods/chat.postMessage#threads.
	ThreadTS string `json:"thread_ts,omitempty"`
	// ReplyBroadcast is used in conjunction with [ThreadTS] and indicates
	// whether the reply should be made visible to everyone in the channel
	// or conversation. Defaults to false.
	ReplyBroadcast bool `json:"reply_broadcast,omitempty"`
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

	Text   string                   `json:"text,omitempty"`
	Blocks []map[string]interface{} `json:"blocks,omitempty"`
	Edited *Edited                  `json:"edited,omitempty"`

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
