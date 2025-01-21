package jira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-webhooks/
type webhook struct {
	ID int `json:"id,omitempty"`
	// Events is the Jira events that trigger the webhook. Valid values:
	// "jira:issue_created", "jira:issue_updated", "jira:issue_deleted",
	// "comment_created", "comment_updated", "comment_deleted",
	// "issue_property_set", "issue_property_deleted".
	Events         []string `json:"events"`
	ExpirationDate string   `json:"expirationDate,omitempty"`
	// FieldIDsFilter is a list of field IDs. When the issue changelog
	// contains any of the fields, the webhook "jira:issue_updated" is sent.
	// If this parameter is not present, the app is notified about all field updates.
	FieldIDsFilter []string `json:"fieldIdsFilter,omitempty"`
	// IssuePropertyKeysFilter is a list of issue property keys.
	// A change of those issue properties triggers the "issue_property_set" or
	// "issue_property_deleted webhooks". If this parameter is not present,
	// the app is notified about all issue property updates.
	IssuePropertyKeysFilter []string `json:"issuePropertyKeysFilter,omitempty"`
	// The JQL filter that specifies which issues the webhook is sent for.
	// Only a subset of JQL can be used. The supported elements are:
	// * Fields: "issueKey", "project", "issuetype", "status", "assignee",
	//   "reporter", "issue.property", and "cf[id]". For custom fields
	//   ("cf[id]"), only the epic label custom field is supported.
	// * Operators: "=", "!=", "IN", and "NOT IN".
	JQLFilter string `json:"jqlFilter"`
}

// checkWebhookPermissions verifies that the OAuth token has the
// necessary Jira permission scopes to manage webhooks. Based on:
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-webhooks/
func checkWebhookPermissions(scopes []string) bool {
	classic := []string{
		"read:jira-work", "manage:jira-webhook",
	}
	granular := []string{
		"read:webhook:jira", "read:jql:jira",
		"read:field:jira", "read:project:jira", "write:webhook:jira",
	}

	allFound := kittehs.All(kittehs.Transform(classic, func(s string) bool {
		return slices.Contains(scopes, s)
	})...)
	if allFound {
		return true
	}

	return kittehs.All(kittehs.Transform(granular, func(s string) bool {
		return slices.Contains(scopes, s)
	})...)
}

type webhookListResponse struct {
	StartAt    int       `json:"startAt"`
	MaxResults int       `json:"maxResults"`
	Total      int       `json:"total"`
	IsLast     bool      `json:"isLast"`
	NextPage   string    `json:"nextPage,omitempty"`
	Values     []webhook `json:"values"`
}

// getWebhook checks whether the given Jira domain already has a
// registered dynamic webhook for this AutoKitteh server. Based on:
// https://developer.atlassian.com/cloud/jira/platform/webhooks/#using-the-rest-api--fetching-registered-webhooks
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-webhooks/#api-rest-api-3-webhook-get
// https://developer.atlassian.com/server/jira/platform/webhooks/
func getWebhook(l *zap.Logger, baseURL, token string) (int, bool) {
	// TODO(ENG-965): Support pagination.
	req, err := http.NewRequest(http.MethodGet, baseURL+"/rest/api/3/webhook", nil)
	if err != nil {
		l.Warn("Failed to construct HTTP request to list Jira webhooks", zap.Error(err))
		return 0, false
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Warn("Failed to list Jira webhooks", zap.Error(err))
		return 0, false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Warn("Failed to read Jira webhooks list response", zap.Error(err))
		return 0, false
	}

	if resp.StatusCode != http.StatusOK {
		l.Warn("Unexpected response to Jira webhooks list request",
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", body),
		)
		return 0, false
	}

	var list webhookListResponse
	if err := json.Unmarshal(body, &list); err != nil {
		l.Warn("Failed to unmarshal Jira webhooks list from JSON",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		return 0, false
	}

	// Finally, filter the results based on the AutoKitteh server address
	// ("GET .../webhook" doesn't show webhook URLs in the response, so
	// we use a trick: we specify the AutoKitteh server address in the
	// JQL filter, without affecting the actual event filtering).
	webhookBase := os.Getenv("WEBHOOK_ADDRESS")
	id := 0
	for _, v := range list.Values {
		if strings.Contains(v.JQLFilter, webhookBase) {
			if id == 0 {
				id = v.ID
			} else {
				// Already found a webhook for this AutoKitteh server.
				// Delete duplicates and return the first one's ID.
				deleteWebhook(l, baseURL, token, v.ID)
			}
		}
	}

	return id, id != 0
}

// deleteWebhook removes the dynamic webhook with the given ID from the Jira domain. Based on:
// https://developer.atlassian.com/cloud/jira/platform/webhooks/#using-the-rest-api--deleting-registered-webhooks
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-webhooks/#api-rest-api-3-webhook-delete
// https://developer.atlassian.com/server/jira/platform/webhooks/
// This function logs errors but doesn't return them, because the only caller is [getWebhook],
// which uses this function to clean up duplicates and doesn't need to know about the errors.
func deleteWebhook(l *zap.Logger, baseURL, oauthToken string, id int) {
	jsonReader := bytes.NewReader([]byte(fmt.Sprintf(`{"webhookIds": [%d]}`, id)))
	req, err := http.NewRequest(http.MethodDelete, baseURL+"/rest/api/3/webhook", jsonReader)
	if err != nil {
		l.Error("Failed to construct HTTP request to delete Jira webhook",
			zap.Int("id", id),
			zap.Error(err),
		)
		return
	}

	req.Header.Set("Authorization", "Bearer "+oauthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Error("Failed to delete Jira webhook",
			zap.Int("id", id),
			zap.Error(err),
		)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		l.Error("Unexpected response to Jira webhook delete request",
			zap.Int("id", id),
			zap.Int("status", resp.StatusCode),
		)
		return
	}

	l.Debug("Deleted a duplicate Jira webhook",
		zap.String("base_url", baseURL),
		zap.Int("id", id),
	)
}

type webhookRegisterRequest struct {
	URL      string    `json:"url"`
	Webhooks []webhook `json:"webhooks"`
}

type webhookRegisterResponse struct {
	Result []webhookRegistrationResult `json:"webhookRegistrationResult"`
}

type webhookRegistrationResult struct {
	CreatedWebhookID int      `json:"createdWebhookId,omitempty"`
	Errors           []string `json:"errors,omitempty"`
}

// registerWebhook creates a new dynamic webhook,
// for 30 days, in the given Jira domain. Based on:
// https://developer.atlassian.com/cloud/jira/platform/webhooks/#using-the-rest-api--registration
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-webhooks/#api-rest-api-3-webhook-post
// https://developer.atlassian.com/server/jira/platform/webhooks/
func registerWebhook(l *zap.Logger, baseURL, token string) (int, bool) {
	webhookBase := os.Getenv("WEBHOOK_ADDRESS")
	r := webhookRegisterRequest{
		URL: fmt.Sprintf("https://%s/jira/webhook", webhookBase),
		Webhooks: []webhook{
			{
				Events: []string{
					"jira:issue_created", "jira:issue_updated", "jira:issue_deleted",
					"comment_created", "comment_updated", "comment_deleted",
					"issue_property_set", "issue_property_deleted",
				},
				// "GET .../webhook" doesn't show webhook URLs in the response,
				// so we use a trick: we specify the AutoKitteh server address in
				// the JQL filter, without affecting the actual event filtering.
				JQLFilter: "project != " + webhookBase,
			},
		},
	}

	body, err := json.Marshal(r)
	if err != nil {
		l.Error("Failed to marshal Jira webhook registration request",
			zap.Any("request", r),
			zap.Error(err),
		)
	}

	jsonReader := bytes.NewReader(body)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/rest/api/3/webhook", jsonReader)
	if err != nil {
		l.Error("Failed to construct HTTP request to register Jira webhook", zap.Error(err))
		return 0, false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Error("Failed to register Jira webhook", zap.Error(err))
		return 0, false
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		l.Error("Failed to read Jira webhook registration response", zap.Error(err))
		return 0, false
	}

	// Error mode 1: based on HTTP status code.
	if resp.StatusCode != http.StatusOK {
		l.Error("Unexpected response to Jira webhook registration request",
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", body),
		)
		return 0, false
	}

	var reg webhookRegisterResponse
	if err := json.Unmarshal(body, &reg); err != nil {
		l.Error("Failed to unmarshal Jira webhook registration result from JSON",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		return 0, false
	}

	// Error mode 2: based on error messages in the parsed JSON response.
	if len(reg.Result) == 0 {
		l.Error("Jira webhook registration result not found", zap.ByteString("body", body))
	}
	if len(reg.Result[0].Errors) > 0 {
		l.Error("Jira webhook registration errors", zap.Strings("errors", reg.Result[0].Errors))
		return 0, false
	}

	// Success.
	l.Info("Registered a new Jira webhook",
		zap.String("base_url", baseURL),
		zap.Int("id", reg.Result[0].CreatedWebhookID),
	)
	return reg.Result[0].CreatedWebhookID, true
}

type webhookRefreshResponse struct {
	ExpirationDate string `json:"expirationDate,omitempty"`
}

// ExtendWebhookLife extends the expiration date of the given webhook ID by 30 days. Based on:
// https://developer.atlassian.com/cloud/jira/platform/webhooks/#using-the-rest-api--refreshing-registered-webhooks
// https://developer.atlassian.com/cloud/jira/platform/rest/v3/api-group-webhooks/#api-rest-api-3-webhook-refresh-put
// https://developer.atlassian.com/server/jira/platform/webhooks/
func ExtendWebhookLife(l *zap.Logger, baseURL, oauthToken string, id int) (time.Time, bool) {
	jsonReader := bytes.NewReader([]byte(fmt.Sprintf(`{"webhookIds": [%d]}`, id)))
	req, err := http.NewRequest(http.MethodPut, baseURL+"/rest/api/3/webhook/refresh", jsonReader)
	if err != nil {
		l.Error("Failed to construct HTTP request to refresh Jira webhook", zap.Error(err))
		return time.Time{}, false
	}

	req.Header.Set("Authorization", "Bearer "+oauthToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Error("Failed to refresh Jira webhook", zap.Error(err))
		return time.Time{}, false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("Failed to read Jira webhook refresh response", zap.Error(err))
		return time.Time{}, false
	}

	if resp.StatusCode != http.StatusOK {
		l.Error("Unexpected response to Jira webhook refresh request",
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", body),
		)
		return time.Time{}, false
	}

	var ref webhookRefreshResponse
	if err := json.Unmarshal(body, &ref); err != nil {
		l.Error("Failed to unmarshal Jira webhook refresh result from JSON",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		return time.Time{}, false
	}

	t, err := time.Parse("2006-01-02T15:04:05.000-0700", ref.ExpirationDate)
	if err != nil {
		l.Error("Failed to parse Jira webhook expiration date",
			zap.String("time", ref.ExpirationDate),
		)
		return utc30Days(), true
	}

	l.Info("Refreshed Jira webhook for new connection",
		zap.String("base_url", baseURL),
		zap.Int("id", id),
		zap.Time("expiration", t),
	)
	return t, true
}

func utc30Days() time.Time {
	return time.Now().UTC().Add(30 * 24 * time.Hour)
}
