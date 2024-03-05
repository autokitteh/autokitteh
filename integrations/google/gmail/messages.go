package gmail

import (
	"context"
	"encoding/base64"
	"errors"

	"google.golang.org/api/gmail/v1"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/batchModify
func (a api) messagesBatchModify(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		ids, addLabelIDs, removeLabelIDs []string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"ids", &ids,
		"add_label_ids?", &addLabelIDs,
		"remove_label_ids?", &removeLabelIDs,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	err = client.Users.Messages.BatchModify("me", &gmail.BatchModifyMessagesRequest{
		Ids:            ids,
		AddLabelIds:    addLabelIDs,
		RemoveLabelIds: removeLabelIDs,
	}).Do()

	// Parse and return the response.
	if err == nil {
		err = errors.New("")
	}
	return sdkvalues.Wrap(err)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/get
func (a api) messagesGet(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	msg, err := client.Users.Messages.Get("me", id).
		Format(format).MetadataHeaders(metadataHeaders...).Do()

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
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/import
func (a api) messagesImport(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		raw                string
		internalDateSource string
		neverMarkSpam      bool
		processForCalendar bool
		deleted            bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"raw", &raw,
		"internal_date_source?", &internalDateSource,
		"never_mark_spam?", &neverMarkSpam,
		"processForCalendar?", &processForCalendar,
		"deleted?", &deleted,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(raw))
	msg := &gmail.Message{Raw: encoded}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	msg, err = client.Users.Messages.Import("me", msg).
		InternalDateSource(internalDateSource).Deleted(deleted).Do()

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
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/insert
func (a api) messagesInsert(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		raw                string
		internalDateSource string
		deleted            bool
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"raw", &raw,
		"internal_date_source?", &internalDateSource,
		"deleted?", &deleted,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(raw))
	msg := &gmail.Message{Raw: encoded}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	msg, err = client.Users.Messages.Insert("me", msg).
		InternalDateSource(internalDateSource).Deleted(deleted).Do()

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
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/list
func (a api) messagesList(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	msgs, err := client.Users.Messages.List("me").
		MaxResults(int64(maxResults)).PageToken(pageToken).Q(q).
		LabelIds(labelIDs...).IncludeSpamTrash(includeSpamTrash).Do()

	// Parse and return the response.
	if msgs == nil {
		msgs = &gmail.ListMessagesResponse{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.ListMessagesResponse `json:"list_messages_response,omitempty"`
		Error                       string `json:"error,omitempty"`
	}{
		msgs,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/modify
func (a api) messagesModify(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	msg, err := client.Users.Messages.Modify("me", id, &gmail.ModifyMessageRequest{
		AddLabelIds:    addLabelIDs,
		RemoveLabelIds: removeLabelIDs,
	}).Do()

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
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/send
func (a api) messagesSend(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
	msg := &gmail.Message{Raw: encoded}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	msg, err = client.Users.Messages.Send("me", msg).Do()

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
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/trash
func (a api) messagesTrash(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
	msg, err := client.Users.Messages.Trash("me", id).Do()

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
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages/untrash
func (a api) messagesUntrash(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
	msg, err := client.Users.Messages.Untrash("me", id).Do()

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
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.messages.attachments/get
func (a api) messagesAttachmentsGet(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		messageID, id string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"message_id", &messageID,
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
	mpb, err := client.Users.Messages.Attachments.Get("me", messageID, id).Do()

	// Parse and return the response.
	if mpb == nil {
		mpb = &gmail.MessagePartBody{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.MessagePartBody `json:"message_part_body,omitempty"`
		Error                  string `json:"error,omitempty"`
	}{
		mpb,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}
