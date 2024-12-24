package jira

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleOAuth receives an inbound redirect request from autokitteh's OAuth
// management service. This request contains an OAuth token (if the OAuth
// flow was successful) and form parameters for debugging and validation
// (either way). If all is well, it saves a new autokitteh connection.
// Either way, it redirects the user to success or failure webpages.
func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Handle errors (e.g. the user didn't authorize us) based on:
	// https://developers.google.com/identity/protocols/oauth2/web-server#handlingresponse
	e := r.FormValue("error")
	if e != "" {
		l.Warn("OAuth redirect reported an error", zap.Error(errors.New(e)))
		c.AbortBadRequest(e)
		return
	}

	_, data, err := sdkintegrations.GetOAuthDataFromURL(r.URL)
	if err != nil {
		l.Warn("Invalid data in OAuth redirect request", zap.Error(err))
		c.AbortBadRequest("invalid data parameter")
		return
	}

	oauthToken := data.Token
	if oauthToken == nil {
		l.Warn("Missing token in OAuth redirect request", zap.Any("data", data))
		c.AbortBadRequest("missing OAuth token")
		return
	}

	url, err := apiBaseURL()
	if err != nil {
		l.Warn("Invalid Atlassian base URL", zap.Error(err))
		c.AbortBadRequest("invalid Atlassian base URL")
		return
	}

	// Test the OAuth token's usability and get authoritative installation details.
	res, err := accessibleOAuth(l, url, oauthToken.AccessToken)
	if err != nil {
		c.AbortBadRequest(err.Error())
		return
	}

	if len(res) > 1 {
		l.Warn("Multiple accessible resources for single OAuth token", zap.Any("resources", res))
		c.AbortBadRequest("multiple Atlassian accessible resources")
		return
	}

	if !checkWebhookPermissions(res[0].Scopes) {
		l.Warn("Insufficient webhook permissions for OAuth token", zap.Any("resources", res))
		c.AbortBadRequest("insufficient webhook permissions")
		return
	}

	// Register a new webhook to receive, parse, and dispatch
	// Jira events, or extend the deadline of an existing one.
	url += "/ex/jira/" + res[0].ID
	t := utc30Days()
	id, ok := getWebhook(l, url, oauthToken.AccessToken)
	if !ok {
		id, ok = registerWebhook(l, url, oauthToken.AccessToken)
		if !ok {
			c.AbortServerError("failed to register webhook")
			return
		}
	} else {
		t, ok = extendWebhookLife(l, url, oauthToken.AccessToken, id)
		if !ok {
			c.AbortServerError("failed to extend webhook life")
			return
		}
	}

	c.Finalize(sdktypes.NewVars(data.ToVars()...).Append(res[0].toVars()...).
		Set(webhookID, fmt.Sprintf("%d", id), false).
		Set(webhookExpiration, t.String(), false))
}

// Determine Jira base URL (to support Jira Data Center, i.e. on-prem).
// TODO(ENG-965): From new-connection form instead of env var.
func apiBaseURL() (string, error) {
	u := os.Getenv("ATLASSIAN_BASE_URL")
	if u == "" {
		u = "https://api.atlassian.com"
	}
	return kittehs.NormalizeURL(u, true)
}

type resource struct {
	ID        string   `json:"id"`
	URL       string   `json:"url"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	AvatarURL string   `json:"avatarUrl"`
}

// accessibleOAuth retrieves the Jira Cloud metadata associated with an
// OAuth token, which is necessary for API calls and webhook events. Based on:
// https://developer.atlassian.com/cloud/jira/platform/oauth-2-3lo-apps/#3--make-calls-to-the-api-using-the-access-token
func accessibleOAuth(l *zap.Logger, baseURL string, token string) ([]resource, error) {
	u := fmt.Sprintf("%s/oauth/token/accessible-resources", baseURL)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		logWarnIfNotNil(l, "Failed to construct HTTP request for OAuth token test", zap.Error(err))
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		logWarnIfNotNil(l, "Failed to request accessible resources for OAuth token", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logWarnIfNotNil(l, "Unexpected response on accessible resources", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("accessible resources: unexpected status code %d", resp.StatusCode)
	}

	var res []resource
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	if len(res) == 0 {
		return nil, errors.New("no accessible resources for OAuth token")
	}

	return res, nil
}

func logWarnIfNotNil(l *zap.Logger, msg string, fields ...zap.Field) {
	if l != nil {
		l.Warn(msg, fields...)
	}
}

func (r resource) toVars() sdktypes.Vars {
	return []sdktypes.Var{
		sdktypes.NewVar(accessID).SetValue(r.ID),
		sdktypes.NewVar(accessURL).SetValue(r.URL),
		sdktypes.NewVar(accessName).SetValue(r.Name),
		sdktypes.NewVar(accessScope).SetValue(fmt.Sprintf("%s", r.Scopes)),
		sdktypes.NewVar(accessAvatarURL).SetValue(r.AvatarURL),
	}
}
