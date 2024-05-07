package users

import (
	"context"
	"net/url"
	"strconv"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

type API struct {
	Vars sdkservices.Vars
}

// TODO: Use HTTP GET instead of POST in all "users" methods.

// TODO: Conversations (https://api.slack.com/methods/users.conversations)

// GetPresence returns a user's current presence status.
//
// Based on: https://api.slack.com/methods/users.getPresence
// See also: https://api.slack.com/docs/presence-and-status
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/users:read
func (a API) GetPresence(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var user string
	if err := sdkmodule.UnpackArgs(args, kwargs, "user?", &user); err != nil {
		return sdktypes.InvalidValue, err
	}
	req := url.Values{}
	req.Set("user", user)

	// Invoke the API method.
	resp := &GetPresenceResponse{}
	err := api.PostForm(ctx, a.Vars, req, resp, "users.getPresence")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// Info returns information about a user based on their Slack user ID, e.g.
// [Profile] details such as their name, email address, and status.
//
// Based on: https://api.slack.com/methods/users.info
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/users:read
//   - https://api.slack.com/scopes/users:read.email
func (a API) Info(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var user string
	var includeLocale bool
	err := sdkmodule.UnpackArgs(args, kwargs,
		"user", &user,
		"include_locale?", &includeLocale,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	req := url.Values{}
	req.Set("user", user)
	if includeLocale {
		req.Set("include_locale", "true")
	}

	// Invoke the API method.
	resp := &InfoResponse{}
	err = api.PostForm(ctx, a.Vars, req, resp, "users.info")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// List returns information about all active and inactive users and apps/bots.
//
// Based on: https://api.slack.com/methods/users.list
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/users:read
//   - https://api.slack.com/scopes/users:read.email
func (a API) List(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var cursor, teamID string
	var limit int
	var includeLocale bool
	err := sdkmodule.UnpackArgs(args, kwargs,
		"cursor?", &cursor,
		"limit?", &limit,
		"include_locale?", &includeLocale,
		"team_id?", &teamID,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	req := url.Values{}
	if cursor != "" {
		req.Set("cursor", cursor)
	}
	if limit > 0 {
		req.Set("limit", strconv.Itoa(limit))
	}
	if includeLocale {
		req.Set("include_locale", "true")
	}
	if teamID != "" {
		req.Set("team_id", teamID)
	}

	// Invoke the API method.
	resp := &ListResponse{}
	err = api.PostForm(ctx, a.Vars, req, resp, "users.list")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// LookupByEmail returns information about a user based on their email address,
// e.g. Slack user and workspace ("team") IDs, [Profile] details, and status.
//
// Based on: https://api.slack.com/methods/users.lookupByEmail
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/users:read
//   - https://api.slack.com/scopes/users:read.email
func (a API) LookupByEmail(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var email string
	if err := sdkmodule.UnpackArgs(args, kwargs, "email", &email); err != nil {
		return sdktypes.InvalidValue, err
	}
	req := url.Values{}
	req.Set("email", email)

	// Invoke the API method.
	resp := &LookupByEmailResponse{}
	err := api.PostForm(ctx, a.Vars, req, resp, "users.lookupByEmail")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}
