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
	ID                       string `json:"id,omitempty"`
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

// Subscribe creates subscriptions for a given list of resources.
// If a subscription already exists for a given resource, we just renew it,
// because Microsoft Graph limits the number of subscriptions per resource to 1.
func Subscribe(ctx context.Context, svc Services, cid sdktypes.ConnectionID, resources []string) []error {
	subs := existingSubscriptions(ctx, svc, cid)

	var errs []error
	for _, r := range resources {
		if _, ok := subs[r]; !ok {
			errs = append(errs, CreateSubscription(ctx, svc, cid, r))
		}
	}

	return errs
}

// existingSubscriptions returns a map of all the existing
// subscriptions for a given connection, keyed by resource name.
func existingSubscriptions(ctx context.Context, svc Services, cid sdktypes.ConnectionID) map[string]*Subscription {
	subs, err := sendRequest(ctx, svc, cid, http.MethodGet, "", nil)
	if err != nil {
		subs = []Subscription{}
	}

	m := make(map[string]*Subscription, len(subs))
	for _, s := range subs {
		m[s.Resource] = &s
	}

	return m
}

// createSubscription creates a subscription to receive change
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

	_, err := sendRequest(ctx, svc, cid, http.MethodPost, "", sub)
	if err == nil {
		svc.Logger.Info("Microsoft Graph subscription created", zap.Any("subscription", sub))
	}
	return err
}

func RenewSubscription(ctx context.Context, svc Services, cid sdktypes.ConnectionID, resource, id string) error {
	sub := &Subscription{
		ExpirationDateTime: expiration(resource),
	}

	_, err := sendRequest(ctx, svc, cid, http.MethodPatch, id, sub)
	if err == nil {
		svc.Logger.Info("Microsoft Graph subscription renewed", zap.Any("subscription", sub))
	}
	return err
}

func DeleteSubscription(ctx context.Context, svc Services, cid sdktypes.ConnectionID, id string) error {
	_, err := sendRequest(ctx, svc, cid, http.MethodDelete, id, nil)
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

// expiration returns a new expiration timestamp for a resource subscription, based on:
// https://learn.microsoft.com/en-us/graph/change-notifications-overview#subscription-lifetime
func expiration(resource string) string {
	t := time.Now().UTC()
	var d time.Duration
	switch {
	case strings.HasPrefix(resource, "/chats"):
		d = 3 * 24 * time.Hour
	case strings.HasPrefix(resource, "/teams"):
		d = 3 * 24 * time.Hour
	default:
		// Safe minimum for unrecognized resources, even though 1 week is preferable.
		d = 6 * time.Hour
	}
	return t.Add(d).Format(time.RFC3339)
}

type subscriptionsList struct {
	Subscriptions []Subscription `json:"value"`
}

// sendRequest sends a subscription-related request to the Microsoft Graph API. If
// this function is successful (error == nil), it updates the input [Subscription] details
// (based on: https://learn.microsoft.com/en-us/graph/change-notifications-delivery-webhooks).
func sendRequest(ctx context.Context, svc Services, cid sdktypes.ConnectionID, httpMethod, subID string, s *Subscription) ([]Subscription, error) {
	l := svc.Logger.With(
		zap.String("http_method", httpMethod),
		zap.String("subscription_id", subID),
	)

	u, err := url.JoinPath(apiURL, subID)
	if err != nil {
		return nil, err
	}

	var r io.Reader = http.NoBody
	if s != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(s); err != nil {
			return nil, err
		}
		r = &buf
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, u, r)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", bearerToken(ctx, l, svc, cid))
	if s != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		l.Warn("MS Graph subscription request failed", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Warn("failed to read MS Graph subscription response", zap.Error(err))
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		if httpMethod == http.MethodGet { // List
			var subs subscriptionsList
			err := json.Unmarshal(body, &subs)
			return subs.Subscriptions, err
		} else { // Create / renew
			return nil, json.Unmarshal(body, s)
		}

	case http.StatusNoContent: // Delete
		return nil, nil

	default:
		l.Warn("MS Graph subscription request failed",
			zap.String("status", resp.Status),
			zap.ByteString("body", body),
		)
		return nil, fmt.Errorf("subscription request failed: %s", resp.Status)
	}
}
