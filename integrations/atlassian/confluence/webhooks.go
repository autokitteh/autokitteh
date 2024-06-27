package confluence

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.jetify.com/typeid"
	"go.uber.org/zap"
)

const (
	restPath = "/wiki/rest/webhooks/1.0/webhook"
)

// https://developer.atlassian.com/cloud/jira/platform/webhooks/#registering-a-webhook-using-the-jira-rest-api--other-integrations-
type webhook struct {
	// Requests.
	Name        string            `json:"name"`
	Description string            `json:"description"`
	URL         string            `json:"url"`
	ExcludeBody bool              `json:"excludeBody"`
	Filters     map[string]string `json:"filters"`
	Events      []string          `json:"events"`
	// https://developer.atlassian.com/cloud/jira/platform/webhooks/#secure-admin-webhooks
	Secret string `json:"secret,omitempty"`

	// Responses.
	Enabled                bool   `json:"enabled,omitempty"`
	Self                   string `json:"self,omitempty"`
	LastUpdatedUser        string `json:"lastUpdatedUser,omitempty"` // UUID
	LastUpdatedDisplayName string `json:"lastUpdatedDisplayName,omitempty"`
	LastUpdated            int    `json:"lastUpdated,omitempty"`
	IsSigned               bool   `json:"isSigned,omitempty"`
}

// getWebhook checks whether the given Confluence domain already has
// a registered dynamic webhook for this AutoKitteh server. Based on:
// https://jira.atlassian.com/browse/CONFCLOUD-36613
// https://developer.atlassian.com/cloud/confluence/modules/webhook/
// https://developer.atlassian.com/server/confluence/webhooks/
// https://confluence.atlassian.com/doc/managing-webhooks-1021225606.html
func getWebhook(l *zap.Logger, base, user, key string) (int, bool) {
	// TODO(ENG-965): Support pagination.
	req, err := http.NewRequest("GET", base+restPath, nil)
	if err != nil {
		l.Warn("Failed to construct HTTP request to list Confluence webhooks", zap.Error(err))
		return 0, false
	}

	req.SetBasicAuth(user, key)

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Warn("Failed to list Confluence webhooks", zap.Error(err))
		return 0, false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Warn("Failed to read Confluence webhooks list response", zap.Error(err))
		return 0, false
	}
	l.Warn("WEBHOOKS", zap.ByteString("body", body)) ///// TODO: REMOVE

	if resp.StatusCode != http.StatusOK {
		l.Warn("Unexpected response to Confluence webhooks list request",
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", body),
		)
		return 0, false
	}

	var list []webhook
	if err := json.Unmarshal(body, &list); err != nil {
		l.Warn("Failed to unmarshal Confluence webhooks list from JSON",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		return 0, false
	}

	// Finally, filter the results based on the AutoKitteh server address.
	webhookBase := os.Getenv("WEBHOOK_ADDRESS")
	url := fmt.Sprintf("https://%s/confluence/webhook", webhookBase)
	for _, w := range list {
		if w.URL == url {
			id, err := extractIDSuffixFromURL(w.Self)
			return id, err == nil
		}
	}

	// No webhook found for this AutoKitteh server.
	return 0, false
}

// registerWebhook creates a new dynamic webhook. Based on:
// https://jira.atlassian.com/browse/CONFCLOUD-36613
// https://developer.atlassian.com/cloud/jira/platform/webhooks/#registering-a-webhook-using-the-jira-rest-api--other-integrations-
// https://developer.atlassian.com/cloud/confluence/modules/webhook/#confluence-webhook-events
func registerWebhook(l *zap.Logger, base, user, key string) (int, string, error) {
	webhookBase := os.Getenv("WEBHOOK_ADDRESS")
	url := fmt.Sprintf("https://%s/confluence/webhook", webhookBase)
	secret := kittehs.Must1(typeid.WithPrefix("")).String()
	r := webhook{
		Name:        "AutoKitteh",
		Description: time.Now().UTC().String(),
		URL:         url,
		ExcludeBody: false,
		Filters:     map[string]string{},
		// https://developer.atlassian.com/cloud/confluence/modules/webhook/
		Events: []string{
			"comment_created",
			"comment_removed",
			"comment_updated",

			"page_created",
			"page_copied",
			"page_moved",
			"page_removed",
			"page_restored",
			"page_trashed",
			"page_unarchived",
			"page_updated",
			"page_viewed",
			"relation_created",
			"relation_deleted",
			"search_performed",

			"user_followed",
			"user_removed",
		},
		Secret: secret,
	}

	body, err := json.Marshal(r)
	if err != nil {
		l.Warn("Failed to marshal Confluence webhook registration request",
			zap.Any("request", r),
			zap.Error(err),
		)
	}

	jsonReader := bytes.NewReader(body)
	req, err := http.NewRequest("POST", base+restPath, jsonReader)
	if err != nil {
		l.Warn("Failed to construct HTTP request to register Confluence webhook", zap.Error(err))
		return 0, "", err
	}

	req.SetBasicAuth(user, key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Warn("Failed to register Confluence webhook", zap.Error(err))
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		l.Warn("Failed to read Confluence webhook registration response", zap.Error(err))
		return 0, "", err
	}

	// Error mode 1: based on HTTP status code.
	if resp.StatusCode != http.StatusCreated {
		l.Warn("Unexpected response to Confluence webhook registration request",
			zap.Int("status", resp.StatusCode),
			zap.ByteString("body", body),
		)
		return 0, "", err
	}

	var reg webhook
	if err := json.Unmarshal(body, &reg); err != nil {
		l.Warn("Failed to unmarshal Confluence webhook registration result from JSON",
			zap.ByteString("body", body),
			zap.Error(err),
		)
		return 0, "", err
	}

	// Error mode 2: based on the content of the parsed JSON response.
	if reg.Self == "" {
		l.Warn("Confluence webhook ID not found", zap.ByteString("body", body))
	}

	id, err := extractIDSuffixFromURL(reg.Self)
	return id, "", err
}

func extractIDSuffixFromURL(url string) (int, error) {
	u := strings.Split(url, "/")
	i, err := strconv.Atoi(u[len(u)-1])
	if err != nil {
		return 0, err
	}
	return i, nil
}
