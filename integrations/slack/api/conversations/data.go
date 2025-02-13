// Package conversations is a lightweight wrapper over the "conversations" methods
// in Slack's Web API: https://api.slack.com/methods?filter=conversations.
package conversations

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api/chat"
)

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
