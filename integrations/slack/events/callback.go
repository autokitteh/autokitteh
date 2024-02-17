package events

// https://api.slack.com/apis/connections/events-api#callback-field
// https://api.slack.com/types/event
type Callback struct {
	// The workspace/team where this event occurred.
	TeamID        string `json:"team_id,omitempty"`
	ContextTeamID string `json:"context_team_id,omitempty"`
	// TODO: context_enterprise_id
	// The application this event is intended for.
	APIAppID string `json:"api_app_id,omitempty"`

	// Typically, this is "event_callback" or "url_verification".
	// See also the event field's "inner event" type.
	Type string `json:"type,omitempty"`

	// Contains the inner set of fields representing the event that's happening.
	Event *Event `json:"event,omitempty"`
	// An installation of your app. Installations are defined by a combination
	// (1 or 2) of the installing Enterprise Grid org, workspace, and user.
	Authorizations []Authorization `json:"authorizations,omitempty"`

	// An identifier for this specific event. Can be used with the
	// https://api.slack.com/methods/apps.event.authorizations.list method to
	// obtain a full list of app installations for which this event is visible.
	// See also: https://api.slack.com/changelog/2020-09-15-events-api-truncate-authed-users#full_list.
	EventContext string `json:"event_context,omitempty"`
	// A unique identifier for this specific event, globally unique across all workspaces.
	EventID string `json:"event_id,omitempty"`
	// The epoch timestamp in seconds indicating when this event was dispatched.
	EventTime int32 `json:"event_time,omitempty"`
}

// https://api.slack.com/apis/connections/events-api#event-type-structure
type Event struct {
	// The specific name of this event - affects parsing.
	Type string `json:"type,omitempty"`
	// The timestamp of what the event describes, which may occur slightly prior
	// to the event being dispatched as described by [EventTS]. The combination
	// of [TS], [TeamID], [UserID], or [ChannelID] is intended to be unique.
	TS string `json:"ts,omitempty"`
	// The timestamp of the event. The combination of [EventTS], [TeamID],
	// [UserID], or [ChannelID] is intended to be unique.
	EventTS string `json:"event_ts,omitempty"`
	// The user ID of to the user that incited this action. Not included in all
	// events as not all events are controlled by users.
	User string `json:"user,omitempty"`
	// Same function as [User], for "channel_member_joined" events.
	Inviter string `json:"inviter,omitempty"`
}

// https://api.slack.com/apis/connections/events-api#authorizations
type Authorization struct {
	EnterpriseID        string `json:"enterprise_id,omitempty"`
	TeamID              string `json:"team_id,omitempty"`
	UserID              string `json:"user_id,omitempty"`
	IsBot               bool   `json:"is_bot,omitempty"`
	IsEnterpriseInstall bool   `json:"is_enterprise_install,omitempty"`
}
