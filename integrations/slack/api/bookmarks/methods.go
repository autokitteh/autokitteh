package bookmarks

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

type API struct {
	Secrets sdkservices.Secrets
	Scope   string
}

// Add a bookmark to a channel.
//
// Based on: https://api.slack.com/methods/bookmarks.add
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/bookmarks:write
func (a API) Add(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := AddRequest{Type: "link"}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel_id", &req.ChannelID,
		"title", &req.Title,
		// "type", &req.Type,
		"link?", &req.Link,
		"emoji?", &req.Emoji,
		"entity_id?", &req.EntityID,
		"parent_id?", &req.ParentID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &AddResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "bookmarks.add")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// Edit a bookmark in a channel.
//
// Based on: https://api.slack.com/methods/bookmarks.edit
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/bookmarks:write
func (a API) Edit(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := EditRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"bookmark_id", &req.BookmarkID,
		"channel_id", &req.ChannelID,
		"emoji?", &req.Emoji,
		"link?", &req.Link,
		"title?", &req.Title,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &EditResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "bookmarks.edit")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// List bookmarks for a channel.
//
// Based on: https://api.slack.com/methods/bookmarks.list
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/bookmarks:read
func (a API) List(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := ListRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel_id", &req.ChannelID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &ListResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "bookmarks.list")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// Remove a bookmark from a channel.
//
// Based on: https://api.slack.com/methods/bookmarks.remove
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/bookmarks:write
func (a API) Remove(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := RemoveRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"bookmark_id", &req.BookmarkID,
		"channel_id", &req.ChannelID,
		"quip_section_id?", &req.QuipSectionID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &RemoveResponse{}
	err = api.PostJSON(ctx, a.Secrets, a.Scope, req, resp, "bookmarks.remove")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}
