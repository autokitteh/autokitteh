// Package users is a lightweight wrapper over the "users" methods
// in Slack's Web API: https://api.slack.com/methods?filter=users.
package users

import (
	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

// -------------------- Requests and responses --------------------

// https://api.slack.com/methods/users.getPresence#examples
type GetPresenceResponse struct {
	api.SlackResponse

	// Presence: "active" or "away".
	Presence string `json:"presence,omitempty"`

	// These fields appear only in responses to self-queries.
	Online          bool `json:"online,omitempty"`
	AutoAway        bool `json:"auto_away,omitempty"`
	ManualAway      bool `json:"manual_away,omitempty"`
	ConnectionCount int  `json:"connection_count,omitempty"`
}

// https://api.slack.com/methods/users.info#examples
type InfoResponse struct {
	api.SlackResponse

	// https://api.slack.com/types/user
	// https://github.com/slackapi/slack-api-specs/blob/master/web-api ("objs_user")
	User *User `json:"user,omitempty"`
}

// https://api.slack.com/methods/users.list#examples
type ListResponse struct {
	api.SlackResponse

	Offset  string `json:"offset,omitempty"`
	CacheTS int    `json:"cache_ts,omitempty"`

	Members []User `json:"members,omitempty"`
}

// https://api.slack.com/methods/users.lookupByEmail#examples
type LookupByEmailResponse struct {
	api.SlackResponse

	// https://api.slack.com/types/user
	// https://github.com/slackapi/slack-api-specs/blob/master/web-api ("objs_user")
	User *User `json:"user,omitempty"`
}

// -------------------- Auxiliary data structures --------------------

// EnterpriseUser contains info related to an Enterprise Grid user.
// See https://api.slack.com/enterprise/grid.
type EnterpriseUser struct {
	EnterpriseID   string `json:"enterprise_id,omitempty"`
	EnterpriseName string `json:"enterprise_name,omitempty"`
	// This user's ID, which might start with "U" or "W": some Grid users have
	// a kind of dual identity - a local, workspace-centric user ID as well as
	// a Grid-wise user ID, called the Enterprise user ID. In most cases these
	// IDs can be used interchangeably, but when it is provided, we strongly
	// recommend using this Enterprise user [ID] over the root level user [ID]
	// field. See also https://api.slack.com/enterprise/grid#user_ids.
	ID      string `json:"id,omitempty"`
	IsAdmin bool   `json:"is_admin,omitempty"`
	IsOwner bool   `json:"is_owner,omitempty"`
	// An array of workspace IDs that are in the Enterprise Grid organization.
	Teams []string `json:"teams,omitempty"`
}

// https://api.slack.com/types/user#profile
// https://github.com/slackapi/slack-api-specs/blob/master/web-api ("objs_user_profile")
type Profile struct {
	// FirstName is the user's first name. The name "slackbot" cannot be used.
	// Updating this will update the first name within [RealName].
	FirstName string `json:"first_name,omitempty"`
	// LastName is the user's last name. The name "slackbot" cannot be used.
	// Updating this will update the second name within [RealName].
	LastName string `json:"last_name,omitempty"`
	// RealName is the user's first and last name. Updating this field will
	// update [FirstName] and [LastName]. If only one name is provided, the
	// value of [LastName] will be cleared.
	RealName string `json:"real_name,omitempty"`
	// RealNameNormalized is the same as [RealName], but with any non-Latin
	// characters filtered out.
	RealNameNormalized string `json:"real_name_normalized,omitempty"`
	// DisplayName is the display name the user has chosen to identify
	// themselves by in their workspace profile. Do not use this field as a
	// unique identifier for a user, as it may change at any time. Instead,
	// use [User.ID] and [User.TeamID] in concert.
	DisplayName string `json:"display_name,omitempty"`
	// DisplayNameNormalized is the same as [DisplayName], but with any
	// non-Latin characters filtered out.
	DisplayNameNormalized string `json:"display_name_normalized,omitempty"`

	// Email is valid, and unique within the workspace. To retrieve it, the
	// https://api.slack.com/scopes/users:read.email scope is required.
	Email     string `json:"email,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Title     string `json:"title,omitempty"`
	Pronouns  string `json:"pronouns,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	// Team ID.
	Team string `json:"team,omitempty"`

	// These 3 fields have a value only if the requested user is a Slack app.
	APIAppID     string `json:"api_app_id,omitempty"`
	BotID        string `json:"bot_id,omitempty"`
	AlwaysActive bool   `json:"always_active,omitempty"`

	// StatusText contains up to 100 characters.
	StatusText string `json:"status_text,omitempty"`
	// StatusTextCanonical is the same as [StatusText],
	// or empty if the text was customized by the user.
	StatusTextCanonical string `json:"status_text_canonical,omitempty"`
	StatusEmoji         string `json:"status_emoji,omitempty"`
	// StatusExpiration is the Unix timestamp of when the status will expire.
	// 0 = custom status that will not expire.
	StatusExpiration int `json:"status_expiration,omitempty"`

	IsCustomImage bool   `json:"is_custom_image,omitempty"`
	ImageOriginal string `json:"image_original,omitempty"`
	Image24       string `json:"image_24,omitempty"`
	Image32       string `json:"image_32,omitempty"`
	Image48       string `json:"image_48,omitempty"`
	Image72       string `json:"image_72,omitempty"`
	Image192      string `json:"image_192,omitempty"`
	Image512      string `json:"image_512,omitempty"`
	Image1024     string `json:"image_1024,omitempty"`
	AvatarHash    string `json:"avatar_hash,omitempty"`

	// TODO: "fields" (https://api.slack.com/methods/users.profile.set#custom_profile).
}

// https://api.slack.com/types/user
// https://github.com/slackapi/slack-api-specs/blob/master/web-api ("objs_user")
type User struct {
	// ID is an identifier for this workspace user. It is unique to the
	// workspace containing the user. Use this field together with [TteamID]
	// as a unique key when storing related data or when specifying the user
	// in API requests. We recommend considering the format of the string to
	// be an opaque value, and not to rely on a particular structure.
	// See also https://api.slack.com/enterprise/grid#user_ids.
	ID       string `json:"id"`
	TeamID   string `json:"team_id"`
	RealName string `json:"real_name,omitempty"`

	// Profile contains the default fields of a user's workspace profile.
	// A user's custom profile fields may be discovered using
	// users.[ProfileGet].
	Profile *Profile `json:"profile,omitempty"`
	// EnterpriseUser contains info related to an Enterprise Grid user.
	// See https://api.slack.com/enterprise/grid.
	EnterpriseUser *EnterpriseUser `json:"enterprise_user,omitempty"`

	// Deleted indicates whether this user has been deactivated.
	Deleted bool `json:"deleted,omitempty"`
	// IsAdmin indicates whether the user is an Admin of the current workspace.
	IsAdmin bool `json:"is_admin,omitempty"`
	// IsAppUser indicates whether the user is an authorized user of the calling app.
	IsAppUser bool `json:"is_app_user,omitempty"`
	// IsBot indicates whether the user is actually a bot user
	// (https://api.slack.com/bot-users). Note that "Slackbot"
	// is special, so [IsBot] will be false for it.
	IsBot            bool `json:"is_bot,omitempty"`
	IsEmailConfirmed bool `json:"is_email_confirmed,omitempty"`
	// IsInvited indicates whether the user has been invited but has not yet signed in.
	// See https://slack.com/intl/en-ie/help/articles/201330256-invite-new-members-to-your-workspace.
	IsInvitedUser bool `json:"is_invited_user,omitempty"`
	// IsOwner indicates whether the user is an Owner of the current workspace.
	IsOwner bool `json:"is_owner,omitempty"`
	// IsPrimaryOwner indicates whether the user is the Primary Owner of the current workspace.
	// See https://slack.com/intl/en-ie/help/articles/360038161033-Understand-the-Primary-Owner-role.
	IsPrimaryOwner bool `json:"is_primary_owner,omitempty"`
	// IsRestricted indicates whether the user is a guest user.
	// See https://slack.com/intl/en-gb/help/articles/202518103-Understand-guest-roles-in-Slack.
	// Use in combination with the [IsUltraRestricted] field to check if the
	// user is a single-channel guest user.
	IsRestricted bool `json:"is_restricted,omitempty"`
	// IsStranger indicates whether the user belongs to a different workspace
	// than the one associated with your app's token, and isn't in any shared
	// channels visible to your app. If false, the user is either from the same
	// workspace as associated with your app's token, or they are from a different
	// workspace, but are in a shared channel that your app has access to. Read our
	// shared channels docs for more details: https://api.slack.com/apis/channels-between-orgs.
	IsStranger bool `json:"is_stranger,omitempty"`
	// IsUltraRestricted indicates whether the user is a single-channel guest.
	// See https://slack.com/intl/en-gb/help/articles/202518103-Understand-guest-roles-in-Slack.
	IsUltraRestricted bool `json:"is_ultra_restricted,omitempty"`

	// A human-readable string for the geographic timezone-related region this
	// user has specified in their account.
	TZ string `json:"tz,omitempty"`
	// Describes the commonly used name of the [TZ] timezone.
	TZLabel string `json:"tz_label,omitempty"`
	// Indicates the number of seconds to offset UTC time by for this user's
	// [TZ]. Changes silently if changed due to daylight savings.
	TZOffset int `json:"tz_offset,omitempty"`

	// A Unix timestamp indicating when the user object was last updated.
	Updated int `json:"updated,omitempty"`

	// Fields that we ignore:

	// "name" - Deprecated. It used to indicate the preferred username.
	// See https://api.slack.com/changelog/2017-09-the-one-about-usernames.

	// "color" - Used in some clients to display a special username color.

	// "locale" - Contains a https://en.wikipedia.org/wiki/IETF_language_tag
	// that represents this user's chosen display language for Slack clients.
	// Useful for https://api.slack.com/start/designing/localizing your apps.
}
