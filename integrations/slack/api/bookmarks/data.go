// Package bookmarks is a lightweight wrapper over the "bookmarks" methods
// in Slack's Web API: https://api.slack.com/methods?filter=bookmarks.
package bookmarks

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/bookmarks.add#args
type AddRequest struct {
	// Channel to add bookmarks in. Required.
	ChannelID string `json:"channel_id"`
	// Title of the bookmark. Required.
	Title string `json:"title"`
	// Type of the bookmark, i.e. "link". Required.
	Type string `json:"type"`
	// Emoji tag to apply to the link.
	Emoji string `json:"emoji,omitempty"`
	// ID of the entity being bookmarked.
	// Only applies to message and file types.
	EntityID string `json:"entity_id,omitempty"`
	// Link to bookmark.
	Link string `json:"link,omitempty"`
	// ID of this bookmark's parent.
	ParentID string `json:"parent_id,omitempty"`
}

// https://api.slack.com/methods/bookmarks.add#examples
type AddResponse struct {
	api.SlackResponse

	Bookmark *Bookmark `json:"bookmark,omitempty"`
}

// https://api.slack.com/methods/bookmarks.edit#args
type EditRequest struct {
	// Bookmark to update. Required.
	BookmarkID string `json:"bookmark_id"`
	// Channel to update bookmark in. Required.
	ChannelID string `json:"channel_id"`
	// Emoji tag to apply to the link.
	Emoji string `json:"emoji,omitempty"`
	// Link to bookmark.
	Link string `json:"link,omitempty"`
	// Title for the bookmark.
	Title string `json:"title,omitempty"`
}

// https://api.slack.com/methods/bookmarks.edit#examples
type EditResponse struct {
	api.SlackResponse

	Bookmark *Bookmark `json:"bookmark,omitempty"`
}

// https://api.slack.com/methods/bookmarks.list#args
type ListRequest struct {
	// Channel to list bookmarks in. Required.
	ChannelID string `json:"channel_id"`
}

// https://api.slack.com/methods/bookmarks.list#examples
type ListResponse struct {
	api.SlackResponse

	Bookmarks []Bookmark `json:"bookmarks,omitempty"`
}

// https://api.slack.com/methods/bookmarks.remove#args
type RemoveRequest struct {
	// Bookmark to remove. Required.
	BookmarkID string `json:"bookmark_id"`
	// Channel to remove bookmark from. Required.
	ChannelID string `json:"channel_id"`
	// Quip section ID to unbookmark.
	QuipSectionID string `json:"quip_section_id,omitempty"`
}

// https://api.slack.com/methods/bookmarks.remove#examples
type RemoveResponse struct {
	api.SlackResponse
}

// -------------------- Auxiliary data structures --------------------

type Bookmark struct {
	ID                  string `json:"id,omitempty"`
	ChannelID           string `json:"channel_id,omitempty"`
	Title               string `json:"title,omitempty"`
	Link                string `json:"link,omitempty"`
	Emoji               string `json:"emoji,omitempty"`
	IconURL             string `json:"icon_url,omitempty"`
	Type                string `json:"type,omitempty"`
	EntityID            string `json:"entity_id,omitempty"`
	Created             int    `json:"date_created,omitempty"`
	Updated             int    `json:"date_updated,omitempty"`
	Rank                string `json:"rank,omitempty"`
	LastUpdatedByUserID string `json:"last_updated_by_user_id,omitempty"`
	LastUpdatedByTeamID string `json:"last_updated_by_team_id,omitempty"`
	ShortcutID          string `json:"shortcut_id,omitempty"`
	AppID               string `json:"app_id,omitempty"`
	AppActionID         string `json:"app_action_id,omitempty"`
}
