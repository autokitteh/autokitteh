package chat

import (
	"context"
	"net/url"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/integrations/slack/api"
)

const (
	DefaultApprovalGreenButton = "Approve"
	DefaultApprovalRedButton   = "Deny"
)

type API struct {
	Vars sdkservices.Vars
}

// Delete an existing message sent by the caller.
//
// Based on: https://api.slack.com/methods/chat.delete
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/chat:write
//   - https://api.slack.com/scopes/chat:write.public (for posting in public
//     channels even when our Slack app isn't a member)
func (a API) Delete(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := DeleteRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"ts", &req.TS,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &DeleteResponse{}
	err = api.PostJSON(ctx, a.Vars, req, resp, "chat.delete")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// GetPermalink generates a permalink URL for a specific extant message.
//
// Based on: https://api.slack.com/methods/chat.getPermalink
//
// Required Slack app scopes: none.
func (a API) GetPermalink(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var channel, messageTS string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &channel,
		"message_ts", &messageTS,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	req := url.Values{}
	req.Set("channel", channel)
	req.Set("message_ts", messageTS)

	// Invoke the API method.
	// TODO: Use HTTP GET instead of POST.
	resp := &GetPermalinkResponse{}
	err = api.PostForm(ctx, a.Vars, req, resp, "chat.getPermalink")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// PostEphemeral sends an ephemeral message to a user in a group/channel
// (visible only to the assigned user). For text formatting tips, see
// https://api.slack.com/reference/surfaces/formatting. This message
// may also contain a rich layout and/or interactive blocks:
// https://api.slack.com/messaging/composing/layouts.
//
// It returns the channel ID and the message's timestamp, but this timestamp
// may not be used for subsequent updates.
//
// https://api.slack.com/methods/chat.postEphemeral#markdown: ephemeral message
// delivery is not guaranteed — the user must be currently active in Slack and
// a member of the specified channel. By nature, ephemeral messages do not
// persist across reloads, desktop and mobile apps, or sessions. Once the
// session is closed, ephemeral messages will disappear and cannot be recovered.
//
// Based on: https://api.slack.com/methods/chat.postEphemeral
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/chat:write
//   - https://api.slack.com/scopes/chat:write.public (for posting in public
//     channels even when our Slack app isn't a member)
func (a API) PostEphemeral(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := PostEphemeralRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"user", &req.User,
		"text", &req.Text,
		"blocks?", &req.Blocks,
		"thread_ts?", &req.ThreadTS,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &PostEphemeralResponse{}
	err = api.PostJSON(ctx, a.Vars, req, resp, "chat.postEphemeral")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// PostMessage sends a message to a user/group/channel. For text formatting
// tips, see https://api.slack.com/reference/surfaces/formatting. This message
// may also contain a rich layout and/or interactive blocks:
// https://api.slack.com/messaging/composing/layouts.
//
// It returns the channel ID, the message's timestamp (for subsequent updates
// or in-thread replies), and a copy of the rendered message.
//
// Based on: https://api.slack.com/methods/chat.postMessage
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/chat:write
//   - https://api.slack.com/scopes/chat:write.public (for posting in public
//     channels even when our Slack app isn't a member)
func (a API) PostMessage(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := PostMessageRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"text?", &req.Text,
		"blocks?", &req.Blocks,
		"thread_ts?", &req.ThreadTS,
		"reply_broadcast?", &req.ReplyBroadcast,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &PostMessageResponse{}
	err = api.PostJSON(ctx, a.Vars, req, resp, "chat.postMessage")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// Update an existing message sent by the caller. For text formatting tips,
// see https://api.slack.com/reference/surfaces/formatting. This message
// may also contain a rich layout and/or interactive blocks:
// https://api.slack.com/messaging/composing/layouts.
//
// It returns the channel ID, the message's timestamp (for subsequent
// updates or in-thread replies), and a copy of the rendered message.
//
// Based on: https://api.slack.com/methods/chat.update
//
// Required Slack app scopes:
//   - https://api.slack.com/scopes/chat:write
//   - https://api.slack.com/scopes/chat:write.public (for posting in
//     public channels even when our Slack app isn't a member)
func (a API) Update(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := UpdateRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"channel", &req.Channel,
		"ts", &req.TS,
		"text?", &req.Text,
		"blocks?", &req.Blocks,
		"reply_broadcast?", &req.ReplyBroadcast,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &UpdateResponse{}
	err = api.PostJSON(ctx, a.Vars, req, resp, "chat.update")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// SendTextMessage sends a message to a user/group/channel. For text formatting
// tips, see https://api.slack.com/reference/surfaces/formatting.
//
// It returns the channel ID, the message's timestamp (for subsequent updates
// or in-thread replies), and a copy of the rendered message.
//
// This is a convenience wrapper over [PostMessage].
func (a API) SendTextMessage(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := PostMessageRequest{}
	err := sdkmodule.UnpackArgs(args, kwargs,
		"target", &req.Channel,
		"text", &req.Text,
		"thread_ts?", &req.ThreadTS,
		"reply_broadcast?", &req.ReplyBroadcast,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	resp := &PostMessageResponse{}
	err = api.PostJSON(ctx, a.Vars, req, resp, "chat.postMessage")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}

// SendApprovalMessage sends an interactive message to a user/group/channel,
// with a short header, a longer message, and 2 buttons. For message formatting
// tips, see https://api.slack.com/reference/surfaces/formatting.
//
// It returns the channel ID, the message's timestamp (for subsequent updates
// or in-thread replies), and a copy of the rendered message. The user's button
// choice will be relayed as an asynchronous interaction event.
//
// This is a convenience wrapper over [PostMessage].
func (a API) SendApprovalMessage(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	req := PostMessageRequest{}
	var message, greenButton, redButton string
	err := sdkmodule.UnpackArgs(args, kwargs,
		"target", &req.Channel,
		"header", &req.Text,
		"message", &message,
		"greenButton?", &greenButton,
		"redButton?", &redButton,
		"thread_ts?", &req.ThreadTS,
		"reply_broadcast?", &req.ReplyBroadcast,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	if greenButton == "" {
		greenButton = DefaultApprovalGreenButton
	}
	if redButton == "" {
		redButton = DefaultApprovalRedButton
	}
	req.Blocks = []Block{
		{
			Type: "header",
			Text: &Text{
				Type:  "plain_text",
				Emoji: true,
				Text:  req.Text,
			},
		},
		{
			Type: "divider",
		},
		{
			Type: "section",
			Text: &Text{
				Type: "mrkdwn",
				Text: message,
			},
		},
		{
			Type: "divider",
		},
		{
			Type: "actions",
			Elements: []Button{
				{
					Type:  "button",
					Style: "primary",
					Text: &Text{
						Type:  "plain_text",
						Emoji: true,
						Text:  greenButton,
					},
					Value:    DefaultApprovalGreenButton,
					ActionID: DefaultApprovalGreenButton,
				},
				{
					Type:  "button",
					Style: "danger",
					Text: &Text{
						Type:  "plain_text",
						Emoji: true,
						Text:  redButton,
					},
					Value:    DefaultApprovalRedButton,
					ActionID: DefaultApprovalRedButton,
				},
			},
		},
	}

	// Invoke the API method.
	resp := &PostMessageResponse{}
	err = api.PostJSON(ctx, a.Vars, req, resp, "chat.postMessage")
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Parse and return the response.
	return sdktypes.WrapValue(resp)
}
