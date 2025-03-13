package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"go.autokitteh.dev/autokitteh/integrations/common"
)

// slackURL is a var and not a const for unit-testing purposes.
var slackURL = "https://slack.com/api"

// get is a helper function to make an HTTP GET request to the Slack API.
// This is used by the [BotsInfo] function during connection initialization.
func get(ctx context.Context, botToken, slackMethod string, jsonResp any) error {
	// Construct the request URL (not using [url.JoinPath] because [slackMethod]
	// may contain query parameters, which [url.JoinPath] will URL-encode).
	u := fmt.Sprintf("%s/%s", slackURL, slackMethod)

	if botToken != "" {
		botToken = "Bearer " + botToken
	}

	resp, err := common.HTTPGet(ctx, u, botToken)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, jsonResp)
}

// Post is a helper function to make an HTTP POST request with a JSON body to the Slack API.
// Use case 1: connection initialization ([AuthTest] and [AppsConnectionsOpen] API calls).
// Use case 2: send updates to Slack webhooks when we finish processing interaction
// events (https://api.slack.com/interactivity/handling#updating_message_response).
func Post(ctx context.Context, botToken, slackMethod string, jsonBody, jsonResp any) error {
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

	if botToken != "" {
		botToken = "Bearer " + botToken
	}

	resp, err := common.HTTPPostJSON(ctx, u, botToken, jsonBody)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp, jsonResp)
}
