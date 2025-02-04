package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	apiURL = "https://graph.microsoft.com/v1.0/subscriptions"

	ChangePath    = "/microsoft/event"
	LifecyclePath = "/microsoft/lifecycle"
)

type Services struct {
	Logger *zap.Logger
	Vars   sdkservices.Vars
	OAuth  sdkservices.OAuth
}

func NewServices(l *zap.Logger, v sdkservices.Vars, o sdkservices.OAuth) Services {
	return Services{Logger: l, Vars: v, OAuth: o}
}

type Subscription struct {
	ChangeType               string `json:"changeType,omitempty"`
	NotificationURL          string `json:"notificationUrl,omitempty"`
	LifecycleNotificationURL string `json:"lifecycleNotificationUrl,omitempty"`
	Resource                 string `json:"resource,omitempty"`
	IncludeResourceData      bool   `json:"includeResourceData,omitempty"`
	EncryptionCertificate    string `json:"encryptionCertificate,omitempty"`
	EncryptionCertificateID  string `json:"encryptionCertificateId,omitempty"`
	ExpirationDateTime       string `json:"expirationDateTime,omitempty"`
	ClientState              string `json:"clientState,omitempty"`
}

// CreateSubscription creates a subscription to receive change
// notifications and lifecycle notifications from Microsoft Graph. Based on:
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks
// https://learn.microsoft.com/en-us/graph/change-notifications-with-resource-data
// https://learn.microsoft.com/en-us/graph/change-notifications-lifecycle-events
func CreateSubscription(ctx context.Context, svc Services, cid sdktypes.ConnectionID, resource string) error {
	changeType := "created,updated,deleted"
	if resource == "/chats/getAllMembers" {
		changeType = "created,deleted"
	}

	changeURL, lifecyleURL := webhookURLs()

	sub := &Subscription{
		ChangeType:               changeType,
		NotificationURL:          changeURL,
		LifecycleNotificationURL: lifecyleURL,
		Resource:                 resource,
		IncludeResourceData:      false, // TODO(INT-233): true,
		// TODO(INT-233): EncryptionCertificate
		// TODO(INT-233): EncryptionCertificateID
		ExpirationDateTime: expiration(resource),
		ClientState:        cid.String(),
	}

	err := sendRequest(ctx, svc, cid, http.MethodPost, "", sub)
	if err == nil {
		svc.Logger.Info("Microsoft Graph subscription created", zap.Any("subscription", sub))
	}
	return err
}

func RenewSubscription(ctx context.Context, svc Services, cid sdktypes.ConnectionID, resource, id string) error {
	sub := &Subscription{
		ExpirationDateTime: expiration(resource),
	}

	err := sendRequest(ctx, svc, cid, http.MethodPatch, id, sub)
	if err == nil {
		svc.Logger.Info("Microsoft Graph subscription renewed", zap.Any("subscription", sub))
	}
	return err
}

func DeleteSubscription(ctx context.Context, svc Services, cid sdktypes.ConnectionID, id string) error {
	err := sendRequest(ctx, svc, cid, http.MethodDelete, id, nil)
	if err == nil {
		svc.Logger.Info("Microsoft Graph subscription deleted", zap.Any("subscription", id))
	}
	return err
}

func webhookURLs() (changeURL, lifecycleURL string) {
	var err error
	changeURL, err = url.JoinPath("https://"+os.Getenv("WEBHOOK_ADDRESS"), ChangePath)
	if err != nil {
		changeURL = ""
	}

	lifecycleURL = strings.ReplaceAll(changeURL, ChangePath, LifecyclePath)
	return
}

// https://learn.microsoft.com/en-us/graph/change-notifications-overview
func expiration(resource string) string {
	t := time.Now().UTC()
	t = t.Add(3 * 24 * time.Hour)
	return t.Format(time.RFC3339)
}

// If this function is successful (error == nil), it updates the input [Subscription] details...
// https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks
func sendRequest(ctx context.Context, svc Services, cid sdktypes.ConnectionID, httpMethod, subID string, s *Subscription) error {
	l := svc.Logger.With(
		zap.String("http_method", httpMethod),
		zap.String("subscription_id", subID),
	)

	u, err := url.JoinPath(apiURL, subID)
	if err != nil {
		return err
	}

	var r io.Reader = http.NoBody
	if s != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(s); err != nil {
			return err
		}
		r = &buf
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, u, r)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", bearerToken(ctx, l, svc, cid))
	if s != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Warn("MS Graph subscription request failed", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Warn("failed to read MS Graph subscription response", zap.Error(err))
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return json.Unmarshal(body, s)
	case http.StatusNoContent:
		return nil
	default:
		l.Warn("MS Graph subscription request failed",
			zap.String("status", resp.Status),
			zap.ByteString("body", body),
		)
		return fmt.Errorf("subscription request failed: %s", resp.Status)
	}
}
