package gmail

import (
	"context"
	"errors"

	"google.golang.org/api/gmail/v1"

	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// https://developers.google.com/gmail/api/reference/rest/v1/users/getProfile
func (a api) getProfile(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	profile, err := client.Users.GetProfile("me").Do()

	// Parse and return the response.
	if profile == nil {
		profile = &gmail.Profile{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.Profile `json:"draft,omitempty"`
		Error          string `json:"error,omitempty"`
	}{
		profile,
		err.Error(),
	}
	return sdktypes.WrapValue(resp)
}

// TODO: https://developers.google.com/gmail/api/reference/rest/v1/users/history/list
func (a api) historyList(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Parse the input arguments.
	var (
		startHistoryID int
		maxResults     int
		pageToken      string
		labelID        string
		historyTypes   []string
	)
	err := sdkmodule.UnpackArgs(args, kwargs,
		"start_history_id", &startHistoryID,
		"max_results?", &maxResults,
		"page_token?", &pageToken,
		"label_id?", &labelID,
		"history_types?", &historyTypes,
	)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	// Invoke the API method.
	client, err := a.gmailClient(ctx)
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	history, err := client.Users.History.List("me").
		MaxResults(int64(maxResults)).PageToken(pageToken).
		StartHistoryId(uint64(startHistoryID)).LabelId(labelID).
		HistoryTypes(historyTypes...).Do()

	// Parse and return the response.
	if history == nil {
		history = &gmail.ListHistoryResponse{}
	}
	if err == nil {
		err = errors.New("")
	}
	resp := struct {
		*gmail.ListHistoryResponse `json:"list_history_response,omitempty"`
		Error                      string `json:"error,omitempty"`
	}{
		history,
		err.Error(),
	}
	return sdktypes.WrapValue(resp)
}
