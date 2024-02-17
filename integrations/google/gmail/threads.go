package gmail

import (
	"context"
	"errors"

	"google.golang.org/api/gmail/v1"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://developers.google.com/gmail/api/reference/rest/v1/users.threads/get
func (a api) threadsGet(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id, format      string
		metadataHeaders []string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
		"format?", &format,
		"metadata_headers?", &metadataHeaders,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return nil, err
	}
	thread, err := client.Users.Threads.Get("me", id).
		Format(format).MetadataHeaders(metadataHeaders...).Do()

	// Parse and return the response.
	if thread == nil {
		thread = &gmail.Thread{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Thread `json:"thread,omitempty"`
		Error         string `json:"error,omitempty"`
	}{
		thread,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.threads/list
func (a api) threadsList(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		maxResults       int
		pageToken, q     string
		labelIDs         []string
		includeSpamTrash bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"max_results?", &maxResults,
		"page_token?", &pageToken,
		"q?", &q,
		"label_ids?", &labelIDs,
		"include_spam_trash?", &includeSpamTrash,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return nil, err
	}
	threads, err := client.Users.Threads.List("me").
		MaxResults(int64(maxResults)).PageToken(pageToken).Q(q).
		LabelIds(labelIDs...).IncludeSpamTrash(includeSpamTrash).Do()

	// Parse and return the response.
	if threads == nil {
		threads = &gmail.ListThreadsResponse{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.ListThreadsResponse `json:"list_threads_response,omitempty"`
		Error                      string `json:"error,omitempty"`
	}{
		threads,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.threads/modify
func (a api) threadsModify(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id             string
		addLabelIDs    []string
		removeLabelIDs []string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
		"add_label_ids?", &addLabelIDs,
		"remove_label_ids?", &removeLabelIDs,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return nil, err
	}
	thread, err := client.Users.Threads.Modify("me", id, &gmail.ModifyThreadRequest{
		AddLabelIds:    addLabelIDs,
		RemoveLabelIds: removeLabelIDs,
	}).Do()

	// Parse and return the response.
	if thread == nil {
		thread = &gmail.Thread{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Thread `json:"thread,omitempty"`
		Error         string `json:"error,omitempty"`
	}{
		thread,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.threads/trash
func (a api) threadsTrash(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return nil, err
	}
	thread, err := client.Users.Threads.Trash("me", id).Do()

	// Parse and return the response.
	if thread == nil {
		thread = &gmail.Thread{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Thread `json:"thread,omitempty"`
		Error         string `json:"error,omitempty"`
	}{
		thread,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.threads/untrash
func (a api) threadsUntrash(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
	)
	if err != nil {
		return nil, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return nil, err
	}
	thread, err := client.Users.Threads.Untrash("me", id).Do()

	// Parse and return the response.
	if thread == nil {
		thread = &gmail.Thread{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Thread `json:"thread,omitempty"`
		Error         string `json:"error,omitempty"`
	}{
		thread,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}
