package gmail

import (
	"context"
	"errors"

	"google.golang.org/api/gmail/v1"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// https://developers.google.com/gmail/api/reference/rest/v1/users.labels/create
func (a api) labelsCreate(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		label *gmail.Label
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"label", &label,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	label, err = client.Users.Labels.Create("me", label).Do()

	// Parse and return the response.
	if label == nil {
		label = &gmail.Label{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Label `json:"label,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		label,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.labels/delete
func (a api) labelsDelete(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
	err = client.Users.Labels.Delete("me", id).Do()

	// Parse and return the response.
	if err == nil {
		err = errors.New("")
	}
	return sdkvalues.Wrap(err)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.labels/get
func (a api) labelsGet(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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
	label, err := client.Users.Labels.Get("me", id).Do()

	// Parse and return the response.
	if label == nil {
		label = &gmail.Label{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Label `json:"label,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		label,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.labels/list
func (a api) labelsList(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	if err := sdkmodule.UnpackArgs(args, kwargs); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	labels, err := client.Users.Labels.List("me").Do()

	// Parse and return the response.
	if labels == nil {
		labels = &gmail.ListLabelsResponse{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.ListLabelsResponse `json:"list_labels_response,omitempty"`
		Error                     string `json:"error,omitempty"`
	}{
		labels,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.labels/patch
func (a api) labelsPatch(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id    string
		label *gmail.Label
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
		"label", &label,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	label, err = client.Users.Labels.Patch("me", id, label).Do()

	// Parse and return the response.
	if label == nil {
		label = &gmail.Label{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Label `json:"label,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		label,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}

// https://developers.google.com/gmail/api/reference/rest/v1/users.labels/update
func (a api) labelsUpdate(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		id    string
		label *gmail.Label
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"id", &id,
		"label", &label,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	label, err = client.Users.Labels.Update("me", id, label).Do()

	// Parse and return the response.
	if label == nil {
		label = &gmail.Label{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Label `json:"label,omitempty"`
		Error        string `json:"error,omitempty"`
	}{
		label,
		err.Error(),
	}
	return sdkvalues.Wrap(resp)
}
