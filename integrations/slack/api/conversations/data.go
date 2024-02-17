// Package conversations is a lightweight wrapper over the "conversations" methods
// in Slack's Web API: https://api.slack.com/methods?filter=conversations.
package conversations

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
	"go.autokitteh.dev/autokitteh/integrations/slack/api/chat"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/conversations.archive#args
type ArchiveRequest struct {
	// ID of the conversation to archive. Required.
	Channel string `json:"channel"`
}

// https://api.slack.com/methods/conversations.archive#examples
type ArchiveResponse struct {
	api.SlackResponse
}

// https://api.slack.com/methods/conversations.close#args
type CloseRequest struct {
	// ID of the conversation to close. Required.
	Channel string `json:"channel"`
}

// https://api.slack.com/methods/conversations.close#examples
type CloseResponse struct {
	api.SlackResponse
}

// https://api.slack.com/methods/conversations.create#args
type CreateRequest struct {
	// Name of the public or private channel to create. Required.
	Name string `json:"name"`
	// IsPrivate indicates whether to create a private
	// channel instead of a public one. Default = false.
	IsPrivate bool `json:"is_private,omitempty"`
	// Team ID to create the channel in. Required if using an org token.
	TeamID string `json:"team_id,omitempty"`
}

// https://api.slack.com/methods/conversations.create#examples
type CreateResponse struct {
	api.SlackResponse

	Channel *Channel `json:"channel,omitempty"`
}

// https://api.slack.com/methods/conversations.history#args
type HistoryRequest struct {
	// Conversation ID to fetch history for. Required.
	Channel string `json:"channel"`

	// Cursor enables pagination through collections of data, based on the
	// [ResponseMetadata.NextCursor] attribute returned in the response to
	// the previous request. The default value ("") fetches the first page
	// of the collection. See https://api.slack.com/docs/pagination and
	// https://api.slack.com/methods/conversations.history#pagination-by-time.
	Cursor string `json:"cursor,omitempty"`
	// Limit is the maximum number of items to return. Fewer than that may
	// be returned, even if the end of the collection hasn't been reached.
	// Default = 100, maximum = 1000. Slack recommends no more than 200.
	Limit int `json:"limit,omitempty"`

	IncludeAllMetadata bool `json:"include_all_metadata,omitempty"`
	// Include messages with [Oldest] or [Latest] timestamps in results?
	// Default = 0. Ignored unless either timestamp is specified.
	Inclusive bool `json:"inclusive,omitempty"`
	// Only messages after this Unix timestamp will be included in results.
	// Default = 0.
	Oldest string `json:"oldest,omitempty"`
	// Only messages before this Unix timestamp will be included in results.
	// Default = current time.
	Latest string `json:"latest,omitempty"`
}

// https://api.slack.com/methods/conversations.history#examples
type HistoryResponse struct {
	api.SlackResponse

	Messages []chat.Message `json:"messages,omitempty"`
	HasMore  bool           `json:"has_more,omitempty"`

	PinCount int `json:"pin_count,omitempty"`
	// Undocumented, always 0?
	ChannelActionsCount int `json:"channel_actions_count,omitempty"`
	// TODO: channel_actions_ts (undocumented, always null?)
}

// https://api.slack.com/methods/conversations.info#examples
type InfoResponse struct {
	api.SlackResponse

	Channel *Channel `json:"channel,omitempty"`
}

// https://api.slack.com/methods/conversations.invite#args
type InviteRequest struct {
	// ID of the public or private channel to invite user(s) to. Required.
	Channel string `json:"channel"`
	// Comma-separated list of user IDs. Up to 1000 users may be listed. Required.
	Users string `json:"users"`
	// When set to true and multiple user IDs are provided, continue inviting
	// the valid ones while disregarding invalid IDs. Defaults to false.
	Force bool `json:"force,omitempty"`
}

// https://api.slack.com/methods/conversations.invite#examples
type InviteResponse struct {
	api.SlackResponse

	Channel *Channel `json:"channel,omitempty"`
	Errors  []Error  `json:"errors,omitempty"`
}

// https://api.slack.com/methods/conversations.list#examples
type ListResponse struct {
	api.SlackResponse

	Channels []Channel `json:"channels,omitempty"`
}

// https://api.slack.com/methods/conversations.members#examples
type MembersResponse struct {
	api.SlackResponse

	Members []string `json:"members,omitempty"`
}

// https://api.slack.com/methods/conversations.open#args
type OpenRequest struct {
	// Resume a conversation by supplying an IM or MPIM's ID.
	// Or provide the "users" field instead.
	Channel string `json:"channel,omitempty"`
	// Comma-separated list of user IDs. If only one user is included, this creates
	// a 1:1 DM. The ordering of the users is preserved whenever a multi-person
	// direct message is returned. Supply a channel when not supplying users.
	Users string `json:"users,omitempty"`
	// Do not create a direct message or multi-person direct message.
	// This is used to see if there is an existing DM or MPDM.
	PreventCreation bool `json:"prevent_creation,omitempty"`
	// Indicates you want the full IM channel definition in the response.
	ReturnIM bool `json:"return_im,omitempty"`
}

// https://api.slack.com/methods/conversations.open#examples
type OpenResponse struct {
	api.SlackResponse

	Channel *Channel `json:"channel,omitempty"`
	Errors  []Error  `json:"errors,omitempty"`
}

// https://api.slack.com/methods/conversations.rename#args
type RenameRequest struct {
	// ID of the conversation (channel) to rename. Required.
	Channel string `json:"channel"`
	// New name for the conversation (channel). Required.
	Name string `json:"name"`
}

// https://api.slack.com/methods/conversations.rename#examples
type RenameResponse struct {
	api.SlackResponse

	Channel *Channel `json:"channel,omitempty"`
}

// https://api.slack.com/methods/conversations.replies#examples
type RepliesResponse struct {
	api.SlackResponse

	Messages []chat.Message `json:"messages,omitempty"`
	HasMore  bool           `json:"has_more,omitempty"`
}

// https://api.slack.com/methods/conversations.setPurpose#args
type SetPurposeRequest struct {
	// ID of the conversation (channel) to set the purpose of. Required.
	Channel string `json:"channel"`
	// A new purpose (description). Required.
	Purpose string `json:"purpose"`
}

// https://api.slack.com/methods/conversations.setPurpose#examples
type SetPurposeResponse struct {
	api.SlackResponse
}

// https://api.slack.com/methods/conversations.setTopic#args
type SetTopicRequest struct {
	// ID of the conversation (channel) to set the topic of. Required.
	Channel string `json:"channel"`
	// A new topic. Required.
	Topic string `json:"topic"`
}

// https://api.slack.com/methods/conversations.setPurpose#examples
type SetTopicResponse struct {
	api.SlackResponse
}

// https://api.slack.com/methods/conversations.unarchive#args
type UnarchiveRequest struct {
	// ID of the conversation to unarchive. Required.
	Channel string `json:"channel,omitempty"`
}

// https://api.slack.com/methods/conversations.unarchive#examples
type UnarchiveResponse struct {
	api.SlackResponse
}

// -------------------- Auxiliary data structures --------------------

// https://api.slack.com/types/conversation
// https://api.slack.com/methods/conversations.info#examples
// https://api.slack.com/methods/conversations.list#examples
// https://github.com/slackapi/slack-api-specs/blob/master/web-api ("objs_conversation")
type Channel struct {
	ID             string   `json:"id,omitempty"`
	Name           string   `json:"name,omitempty"`
	NameNormalized string   `json:"name_normalized,omitempty"`
	PreviousNames  []string `json:"previous_name,omitempty"`
	Creator        string   `json:"creator,omitempty"`
	User           string   `json:"user,omitempty"`

	IsMember     bool `json:"is_member,omitempty"`
	IsReadOnly   bool `json:"is_read_only,omitempty"`
	IsThreadOnly bool `json:"is_thread_only,omitempty"`

	Topic       *Topic   `json:"topic,omitempty"`
	Purpose     *Purpose `json:"purpose,omitempty"`
	LastRead    string   `json:"last_read,omitempty"`
	UnreadCount int      `json:"unread_count,omitempty"`
	// A count of messages that the calling user has yet to read that matter to
	// them (excludes things like join/leave messages).
	UnreadCountDisplay int `json:"unread_count_display,omitempty"`

	IsArchived bool `json:"is_archived,omitempty"`
	IsChannel  bool `json:"is_channel,omitempty"`
	IsFrozen   bool `json:"is_frozen,omitempty"`
	IsGeneral  bool `json:"is_general,omitempty"`
	IsGroup    bool `json:"is_group,omitempty"`
	IsIM       bool `json:"is_im,omitempty"`
	IsMPIM     bool `json:"is_mpim,omitempty"`
	IsOpen     bool `json:"is_open,omitempty"`
	IsPrivate  bool `json:"is_private,omitempty"`

	IsShared           bool `json:"is_shared,omitempty"`
	IsOrgShared        bool `json:"is_org_shared,omitempty"`
	IsExtShared        bool `json:"is_ext_shared,omitempty"`
	IsPendingExtShared bool `json:"is_pending_ext_shared,omitempty"`

	ContextTeamID           string   `json:"context_team_id,omitempty"`
	SharedTeamIDs           []string `json:"shared_team_ids,omitempty"`
	PendingConnectedTeamIDs []string `json:"pending_connected_team_ids,omitempty"`

	Created  int `json:"created,omitempty"`
	Updated  int `json:"updated,omitempty"`
	Unlinked int `json:"unlinked,omitempty"`

	Latest *chat.Message `json:"latest,omitempty"`

	Locale     string  `json:"locale,omitempty"`
	NumMembers int     `json:"num_members,omitempty"`
	Priority   float32 `json:"priority,omitempty"`

	// TODO: "parent_conversation"
	// TODO: "pending_shared"
	// TODO: "priority"
}

type Error struct {
	User  string `json:"user,omitempty"`
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type Purpose struct {
	Value   string `json:"value,omitempty"`
	Creator string `json:"creator,omitempty"`
	LastSet int    `json:"last_set,omitempty"`
}

type Topic struct {
	Value   string `json:"value,omitempty"`
	Creator string `json:"creator,omitempty"`
	LastSet int    `json:"last_set,omitempty"`
}
