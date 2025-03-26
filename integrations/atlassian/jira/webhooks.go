package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
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
func getWebhook(ctx context.Context, l *zap.Logger, baseURL, token string) (int, bool) {
	// TODO(ENG-965): Support pagination.
	u := baseURL + "/rest/api/3/webhook"
	resp, err := common.HTTPGet(ctx, u, "Bearer "+token)
	if err != nil {
		l.Warn("Failed to request Jira webhook", zap.Error(err))
		return 0, false
	}

	var list webhookListResponse
	if err := json.Unmarshal(resp, &list); err != nil {
		l.Warn("Failed to unmarshal Jira webhooks list from JSON",
			zap.ByteString("body", resp),
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
				deleteWebhook(ctx, l, baseURL, token, v.ID)
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
func deleteWebhook(ctx context.Context, l *zap.Logger, baseURL, oauthToken string, id int) {
	ctx, cancel := context.WithTimeout(ctx, common.HTTPTimeout)
	defer cancel()

	u := baseURL + "/rest/api/3/webhook"
	req := fmt.Sprintf(`{"webhookIds": [%d]}`, id)
	if _, err := common.HTTPDeleteJSON(ctx, u, "Bearer "+oauthToken, req); err != nil {
		l.Error("failed to delete Jira webhook", zap.Int("id", id), zap.Error(err))
		return
	}

	l.Debug("Deleted a duplicate Jira webhook",
		zap.String("base_url", baseURL), zap.Int("id", id),
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
func registerWebhook(ctx context.Context, l *zap.Logger, baseURL, token string) (int, bool) {
	webhookBase := os.Getenv("WEBHOOK_ADDRESS")
	req := webhookRegisterRequest{
		URL: fmt.Sprintf("https://%s/jira/webhook", webhookBase),
		Webhooks: []webhook{
			{
				Events: []string{
					"jira:issue_created", "jira:issue_updated", "jira:issue_deleted",
					"comment_created", "comment_updated", "comment_deleted",
				},
				// "GET .../webhook" doesn't show webhook URLs in the response,
				// so we use a trick: we specify the AutoKitteh server address in
				// the JQL filter, without affecting the actual event filtering.
				JQLFilter: "project != " + webhookBase,
			},
		},
	}

	u := baseURL + "/rest/api/3/webhook"
	resp, err := common.HTTPPostJSON(ctx, u, "Bearer "+token, req)
	if err != nil {
		l.Error("failed to register Jira webhook", zap.Error(err))
		return 0, false
	}

	var reg webhookRegisterResponse
	if err := json.Unmarshal(resp, &reg); err != nil {
		l.Error("failed to unmarshal Jira webhook registration result from JSON",
			zap.ByteString("body", resp),
			zap.Error(err),
		)
		return 0, false
	}

	if len(reg.Result) == 0 {
		l.Error("Jira webhook registration result not found", zap.ByteString("body", resp))
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
func ExtendWebhookLife(ctx context.Context, l *zap.Logger, baseURL, oauthToken string, id int) (time.Time, bool) {
	u := baseURL + "/rest/api/3/webhook/refresh"
	req := fmt.Sprintf(`{"webhookIds": [%d]}`, id)
	resp, err := common.HTTPPutJSON(ctx, u, "Bearer "+oauthToken, req)
	if err != nil {
		l.Error("failed to refresh Jira webhook", zap.Error(err))
		return time.Time{}, false
	}

	var ref webhookRefreshResponse
	if err := json.Unmarshal(resp, &ref); err != nil {
		l.Error("failed to unmarshal Jira webhook refresh result from JSON",
			zap.ByteString("body", resp),
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
