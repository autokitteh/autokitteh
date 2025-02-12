package confluence

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	Description string            `json:"description,omitempty"`
	URL         string            `json:"url"`
	ExcludeBody bool              `json:"excludeBody,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	Events      []string          `json:"events"`
	// https://developer.atlassian.com/cloud/jira/platform/webhooks/#secure-admin-webhooks
	// TODO(ENG-1081): Empirically, Confluence doesn't recognize this!
	Secret string `json:"secret"`

	// Responses.
	Enabled                bool   `json:"enabled,omitempty"`
	Self                   string `json:"self,omitempty"`
	LastUpdatedUser        string `json:"lastUpdatedUser,omitempty"`
	LastUpdatedDisplayName string `json:"lastUpdatedDisplayName,omitempty"`
	LastUpdated            int    `json:"lastUpdated,omitempty"`
	IsSigned               bool   `json:"isSigned,omitempty"`
}

// https://developer.atlassian.com/cloud/confluence/modules/webhook/
// https://confluence.atlassian.com/doc/managing-webhooks-1021225606.html
var webhookEvents = map[string][]string{
	"added": {
		"label_added",
	},
	"archived": {
		"attachment_archived",
		"page_archived",
	},
	"copied": {
		"page_copied",
	},
	"created": {
		"attachment_created",
		"blog_created",
		"blueprint_page_created",
		"comment_created",
		"content_created",
		"group_created",
		"label_created",
		"page_created",
		"relation_created",
		"space_created",
	},
	"deleted": {
		"label_deleted",
		"relation_deleted",
	},
	"moved": {
		"page_moved",
	},
	"removed": {
		"attachment_removed",
		"blog_removed",
		"comment_removed",
		"content_removed",
		"group_removed",
		"label_removed",
		"page_removed",
		"space_removed",
		"user_removed",
	},
	"trashed": {
		"attachment_trashed",
		"blog_trashed",
		"content_trashed",
		"page_trashed",
	},
	"updated": {
		"attachment_updated",
		"blog_updated",
		"comment_updated",
		"content_updated",
		"page_updated",
		"space_logo_updated",
		"space_updated",
	},
}

// getWebhook checks whether the given Confluence domain already has
// a registered dynamic webhooks for this AutoKitteh server. Based on:
// https://jira.atlassian.com/browse/CONFCLOUD-36613
// https://developer.atlassian.com/cloud/confluence/modules/webhook/
// https://developer.atlassian.com/server/confluence/webhooks/
// https://confluence.atlassian.com/doc/managing-webhooks-1021225606.html
func getWebhook(ctx context.Context, l *zap.Logger, base, user, key, category string) (int, bool) {
	// TODO(ENG-965): Support pagination.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+restPath, nil)
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
	url := fmt.Sprintf("https://%s/confluence/webhook/%s", webhookBase, category)
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
func registerWebhook(ctx context.Context, l *zap.Logger, base, user, key, category string) (int, string, error) {
	l = l.With(zap.String("category", category))

	webhookBase := os.Getenv("WEBHOOK_ADDRESS")
	url := fmt.Sprintf("https://%s/confluence/webhook/%s", webhookBase, category)
	secret := typeid.Must(typeid.WithPrefix("")).String()
	r := webhook{
		Name:        "AutoKitteh",
		Description: time.Now().UTC().String(),
		URL:         url,
		Events:      webhookEvents[category],
		Secret:      secret,
	}

	body, err := json.Marshal(r)
	if err != nil {
		l.Warn("Failed to marshal Confluence webhook registration request",
			zap.Any("request", r),
			zap.Error(err),
		)
		return 0, "", err
	}

	jsonReader := bytes.NewReader(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+restPath, jsonReader)
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
		s := strings.TrimSpace(string(body))
		s = s[:min(len(s), 256)]
		if s == "" {
			s = "no error message"
		}
		return 0, "", errors.New(s)
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
		return 0, "", errors.New("no webhook ID in response")
	}

	// Success.
	id, err := extractIDSuffixFromURL(reg.Self)
	l.Info("Registered Confluence events webhook", zap.Int("id", id))
	return id, secret, err
}

func extractIDSuffixFromURL(url string) (int, error) {
	u := strings.Split(url, "/")
	i, err := strconv.Atoi(u[len(u)-1])
	if err != nil {
		return 0, err
	}
	return i, nil
}

func deleteWebhook(ctx context.Context, l *zap.Logger, base, user, key string, id int) error {
	url := fmt.Sprintf("%s%s/%d", base, restPath, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		l.Error("Failed to construct HTTP request to delete Confluence webhook", zap.Error(err))
		return err
	}

	req.SetBasicAuth(user, key)

	resp, err := httpClient.Do(req)
	if err != nil {
		l.Error("Failed to delete Confluence webhook", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		l.Error("Unexpected response to Confluence webhook deletion request",
			zap.Int("status", resp.StatusCode),
		)
		return fmt.Errorf("existing webhook deletion failed: %d", resp.StatusCode)
	}

	return nil
}
