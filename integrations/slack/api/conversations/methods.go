package conversations

import (
	"context"
	"net/url"
	"strconv"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

type API struct {
	Secrets sdkservices.Secrets
	Scope   string
}

// Archive a conversation (channel).
//
// Based on: https://api.slack.com/methods/conversations.archive
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
func (a API) Archive(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := ArchiveRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &ArchiveResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.archive")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Close a direct message or multi-person direct message.
//
// Based on: https://api.slack.com/methods/conversations.close
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
func (a API) Close(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := CloseRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &CloseResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.close")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Create initiates a public or private channel-based conversation.
//
// Based on: https://api.slack.com/methods/conversations.create
//
// Channel names may only contain lowercase letters, numbers,
// hyphens, and underscores, and must be 80 characters or less.
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
func (a API) Create(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := CreateRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"name", &req.Name,
		"is_private?", &req.IsPrivate,
		"team_id?", &req.TeamID,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &CreateResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.create")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// History returns the history of a conversation's (channel's)
// messages and events.
//
// Based on: https://api.slack.com/methods/conversations.history
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:history
//   - https://api.slack.com/scopes/groups:history
//   - https://api.slack.com/scopes/im:history
//   - https://api.slack.com/scopes/mpim:history
func (a API) History(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := HistoryRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"cursor?", &req.Cursor,
		"limit?", &req.Limit,
		"include_all_metadata?", &req.IncludeAllMetadata,
		"inclusive?", &req.Inclusive,
		"oldest?", &req.Oldest,
		"latest?", &req.Latest,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &HistoryResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.history")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Info returns information about a conversation (channel).
//
// Based on: https://api.slack.com/methods/conversations.info
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:read
//   - https://api.slack.com/scopes/groups:read
//   - https://api.slack.com/scopes/im:read
//   - https://api.slack.com/scopes/mpim:read
func (a API) Info(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		channel                          string
		includeLocale, includeNumMembers bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &channel,
		"include_locale?", &includeLocale,
		"include_num_members?", &includeNumMembers,
	)
	if err != nil {
		return nil, err
	}
	req := url.Values{}
	req.Set("channel", channel)
	if includeLocale {
		req.Set("include_locale", "true")
	}
	if includeNumMembers {
		req.Set("include_num_members", "true")
	}

	// Invoke the API method.
	// TODO: Use HTTP GET instead of POST.
	resp := &InfoResponse{}
	err = api.PostForm(ctx, a.Secrets, a.Scope, req, resp, "conversations.info")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Invite invites users to a channel.
//
// Based on: https://api.slack.com/methods/conversations.invite
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/channels:write.invites
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/groups:write.invites
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
//   - https://api.slack.com/scopes/mpim:write.invites
func (a API) Invite(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := InviteRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"users", &req.Users,
		"force?", &req.Force,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &InviteResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.invite")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// List lists all the channels in a Slack team.
//
// Based on: https://api.slack.com/methods/conversations.list
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:read
//   - https://api.slack.com/scopes/groups:read
//   - https://api.slack.com/scopes/im:read
//   - https://api.slack.com/scopes/mpim:read
func (a API) List(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		cursor, teamID, types string
		limit                 int
		excludeArchived       bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"cursor?", &cursor,
		"limit?", &limit,
		"exclude_archived?", &excludeArchived,
		"team_id?", &teamID,
		"types?", &types,
	)
	if err != nil {
		return nil, err
	}
	req := url.Values{}
	if cursor != "" {
		req.Set("cursor", cursor)
	}
	if limit > 0 {
		req.Set("limit", strconv.Itoa(limit))
	}
	if excludeArchived {
		req.Set("exclude_archived", "true")
	}
	if teamID != "" {
		req.Set("team_id", teamID)
	}
	if types != "" {
		req.Set("types", types)
	}

	// Invoke the API method.
	// TODO: Use HTTP GET instead of POST.
	resp := &ListResponse{}
	err = api.PostForm(ctx, a.Secrets, a.Scope, req, resp, "conversations.list")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Members retrieves all the members of a conversation (channel).
//
// Based on: https://api.slack.com/methods/conversations.members
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:read
//   - https://api.slack.com/scopes/groups:read
//   - https://api.slack.com/scopes/im:read
//   - https://api.slack.com/scopes/mpim:read
func (a API) Members(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		channel, cursor string
		limit           int
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &channel,
		"cursor?", &cursor,
		"limit?", &limit,
	)
	if err != nil {
		return nil, err
	}
	req := url.Values{}
	req.Set("channel", channel)
	if cursor != "" {
		req.Set("cursor", cursor)
	}
	if limit > 0 {
		req.Set("limit", strconv.Itoa(limit))
	}

	// Invoke the API method.
	// TODO: Use HTTP GET instead of POST.
	resp := &MembersResponse{}
	err = api.PostForm(ctx, a.Secrets, a.Scope, req, resp, "conversations.members")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Open opens or resumes a direct message or multi-person direct message.
//
// Based on: https://api.slack.com/methods/conversations.open
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
func (a API) Open(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := OpenRequest{ReturnIM: true}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel?", &req.Channel,
		"users?", &req.Users,
		"prevent_creation?", &req.PreventCreation,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &OpenResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.open")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Rename a conversation (channel).
//
// Based on: https://api.slack.com/methods/conversations.rename
//
// Channel names may only contain lowercase letters, numbers,
// hyphens, and underscores, and must be 80 characters or less.
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
func (a API) Rename(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := RenameRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"name", &req.Name,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &RenameResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.rename")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Replies retrieves a thread of messages posted to a conversation (channel).
//
// Based on: https://api.slack.com/methods/conversations.replies
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:history
//   - https://api.slack.com/scopes/groups:history
//   - https://api.slack.com/scopes/im:history
//   - https://api.slack.com/scopes/mpim:history
func (a API) Replies(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		channel, ts, cursor, oldest, latest string
		limit                               int
		includeAllMetadata, inclusive       bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &channel,
		"ts", &ts,
		"cursor?", &cursor,
		"limit?", &limit,
		"include_all_metadata?", &includeAllMetadata,
		"inclusive?", &inclusive,
		"oldest?", &oldest,
		"latest?", &latest,
	)
	if err != nil {
		return nil, err
	}
	req := url.Values{}
	req.Set("channel", channel)
	req.Set("ts", ts)
	if cursor != "" {
		req.Set("cursor", cursor)
	}
	if limit > 0 {
		req.Set("limit", strconv.Itoa(limit))
	}
	if includeAllMetadata {
		req.Set("include_all_metadata", "true")
	}
	if inclusive {
		req.Set("inclusive", "true")
	}
	if oldest != "" {
		req.Set("oldest", oldest)
	}
	if latest != "" {
		req.Set("latest", latest)
	}

	// Invoke the API method.
	// TODO: Use HTTP GET instead of POST.
	resp := &RepliesResponse{}
	err = api.PostForm(ctx, a.Secrets, a.Scope, req, resp, "conversations.replies")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// SetPurpose sets the purpose (description) for a conversation (channel).
//
// Based on: https://api.slack.com/methods/conversations.setPurpose
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/channels:write.topic
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/groups:write.topic
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
//   - https://api.slack.com/scopes/mpim:write.topic
func (a API) SetPurpose(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := SetPurposeRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"purpose", &req.Purpose,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &SetPurposeResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.setPurpose")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// SetTopic sets the topic for a conversation (channel).
//
// Based on: https://api.slack.com/methods/conversations.setTopic
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/channels:write.topic
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/groups:write.topic
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
//   - https://api.slack.com/scopes/mpim:write.topic
func (a API) SetTopic(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := SetTopicRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"topic", &req.Topic,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &SetTopicResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.setTopic")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}

// Unarchive reverses conversation (channel) archival.
//
// Based on: https://api.slack.com/methods/conversations.unarchive
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/channels:manage
//   - https://api.slack.com/scopes/groups:write
//   - https://api.slack.com/scopes/im:write
//   - https://api.slack.com/scopes/mpim:write
func (a API) Unarchive(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := UnarchiveRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	resp := &UnarchiveResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "conversations.unarchive")
	if err != nil {
		return nil, err
	}

	// Parse and return the response.
	return sdkvalues.Wrap(resp)
}
