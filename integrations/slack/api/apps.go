package api

import (
	"context"
	"errors"
	"net/http"
)

// https://api.slack.com/methods/apps.connections.open#examples
type AppsConnectionsOpenResponse struct {
	SlackResponse

	URL string `json:"url"`
}

// AppsConnectionsOpen generates a temporary WebSocket URL for a Socket Mode
// app, to check the usability of an app-level token provided by the user.
// Based on: https://api.slack.com/methods/apps.connections.open
// Required Slack app scope: https://api.slack.com/scopes/connections:write
func AppsConnectionsOpen(ctx context.Context, appToken string) (*AppsConnectionsOpenResponse, error) {
	resp := &AppsConnectionsOpenResponse{}
	if err := Post(ctx, appToken, "apps.connections.open", http.NoBody, resp); err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, errors.New(resp.Error)
	}
	return resp, nil
}
