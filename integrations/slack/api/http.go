package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	headerAccept        = "Accept"
	headerAuthorization = "Authorization"
	HeaderContentType   = "Content-Type"

	ContentTypeJSON            = "application/json"                // Accept
	ContentTypeJSONCharsetUTF8 = "application/json; charset=utf-8" // Content-Type
	ContentTypeForm            = "application/x-www-form-urlencoded"

	timeout = 3 * time.Second
)

// slackURL is a var and not a const for unit-testing purposes.
var slackURL = "https://slack.com/api"

// get is a helper function to make an HTTP GET request to the Slack API.
// This is used by the [BotsInfo] function during connection initialization.
func get(ctx context.Context, botToken, slackMethod string, jsonResp any) error {
	// Construct the request URL (not using [url.JoinPath] because [slackMethod]
	// may contain query parameters, which [url.JoinPath] will URL-encode).
	u := fmt.Sprintf("%s/%s", slackURL, slackMethod)

	// Construct the request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return err
	}

	req.Header.Set(headerAccept, ContentTypeJSON)
	if botToken != "" {
		req.Header.Set(headerAuthorization, "Bearer "+botToken)
	}

	// Send the request to the server.
	c := &http.Client{Timeout: timeout}
	httpResp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	// Parse the HTTP response.
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", httpResp.StatusCode, http.StatusText(httpResp.StatusCode))
	}

	b, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, jsonResp)
}

// Post is a helper function to make an HTTP POST request with a JSON body to the Slack API.
// Use case 1: connection initialization ([AuthTest] and [AppsConnectionsOpen] API calls).
// Use case 2: send updates to Slack webhooks when we finish processing interaction
// events (https://api.slack.com/interactivity/handling#updating_message_response).
func Post(ctx context.Context, botToken, slackMethod string, body, resp any) error {
	// Construct the request URL: if slackMethod is a full URL (a callback URL
	// provided by a Slack event) then don't prepend the Slack API's URL.
	u := slackMethod
	if !strings.HasPrefix(u, "https://") {
		var err error
		u, err = url.JoinPath(slackURL, slackMethod)
		if err != nil {
			return err
		}
	}

	// Construct the request body.
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}

	// Construct the request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.Header.Set(HeaderContentType, ContentTypeJSONCharsetUTF8)
	req.Header.Set(headerAccept, ContentTypeJSON)
	if botToken != "" {
		req.Header.Set(headerAuthorization, "Bearer "+botToken)
	}

	// Send request to server.
	c := &http.Client{Timeout: timeout}
	httpResp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer httpResp.Body.Close()

	// Parse the HTTP response.
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("%d %s", httpResp.StatusCode, http.StatusText(httpResp.StatusCode))
	}

	b, err = io.ReadAll(httpResp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, resp)
}
