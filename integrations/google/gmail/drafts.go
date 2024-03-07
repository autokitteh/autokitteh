package gmail

import (
	"context"
	"encoding/base64"
	"errors"

	"google.golang.org/api/gmail/v1"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/create
func (a api) draftsCreate(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		raw string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"raw", &raw,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(raw))
	draft := &gmail.Draft{Message: &gmail.Message{Raw: encoded}}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	draft, err = client.Users.Drafts.Create("me", draft).Do()

	// Parse and return the response.
	if draft == nil {
		draft = &gmail.Draft{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Draft `json:"draft,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		draft,
		err.Error(),
	}
	return sdktypes.WrapValue(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/delete
func (a api) draftsDelete(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	err = client.Users.Drafts.Delete("me", id).Do()

	// Parse and return the response.
	if err == nil {
		err = errors.New("")
	}
	return sdktypes.WrapValue(err)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/get
func (a api) draftsGet(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id, format string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
		"format?", &format,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	draft, err := client.Users.Drafts.Get("me", id).Format(format).Do()

	// Parse and return the response.
	if draft == nil {
		draft = &gmail.Draft{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Draft `json:"draft,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		draft,
		err.Error(),
	}
	return sdktypes.WrapValue(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/list
func (a api) draftsList(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		maxResults       int
		pageToken, q     string
		includeSpamTrash bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"max_results?", &maxResults,
		"page_token?", &pageToken,
		"q?", &q,
		"include_spam_trash?", &includeSpamTrash,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	drafts, err := client.Users.Drafts.List("me").
		MaxResults(int64(maxResults)).PageToken(pageToken).
		Q(q).IncludeSpamTrash(includeSpamTrash).Do()

	// Parse and return the response.
	if drafts == nil {
		drafts = &gmail.ListDraftsResponse{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.ListDraftsResponse `json:"list_drafts_response,omitempty"`
		Error                     string `json:"error,omitempty"`
	}{
		drafts,
		err.Error(),
	}
	return sdktypes.WrapValue(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/send
func (a api) draftsSend(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		raw string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"raw", &raw,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(raw))
	draft := &gmail.Draft{Message: &gmail.Message{Raw: encoded}}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	msg, err := client.Users.Drafts.Send("me", draft).Do()

	// Parse and return the response.
	if msg == nil {
		msg = &gmail.Message{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Message `json:"message,omitempty"`
		Error          string `json:"error,omitempty"`
	}{
		msg,
		err.Error(),
	}
	return sdktypes.WrapValue(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.drafts/update
func (a api) draftsUpdate(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id, raw string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
		"raw", &raw,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(raw))
	draft := &gmail.Draft{Message: &gmail.Message{Raw: encoded}}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	draft, err = client.Users.Drafts.Update("me", id, draft).Do()

	// Parse and return the response.
	if draft == nil {
		draft = &gmail.Draft{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Draft `json:"draft,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		draft,
		err.Error(),
	}
	return sdktypes.WrapValue(resp)
}
