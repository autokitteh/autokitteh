package google

import (
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// calWebhookPath is the URL path to receive incoming Google Calendar push notifications.
	calWebhookPath = "/googlecalendar/notif"
	// driveWebhookPath is the URL path to receive incoming Google Drive push notifications.
	driveWebhookPath = "/googledrive/notif"
	// formsWebhookPath is the URL path to receive incoming Google Forms push notifications.
	formsWebhookPath = "/googleforms/notif"
	// gmailWebhookPath is the URL path to receive incoming Gmail push notifications.
	gmailWebhookPath = "/gmail/notif"
)

// Start initializes all the HTTP handlers of all the Google integrations.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, m *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.DispatchFunc) {
	common.ServeStaticUI(m, desc, static.GoogleWebContent)
	uiPath := fmt.Sprintf("GET %s/", desc.ConnectionURL().Path)

	urlPath := strings.ReplaceAll(uiPath, "google", "gmail")
	m.Auth.Handle(urlPath, http.FileServer(http.FS(static.GmailWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlecalendar")
	m.Auth.Handle(urlPath, http.FileServer(http.FS(static.GoogleCalendarWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlechat")
	m.Auth.Handle(urlPath, http.FileServer(http.FS(static.GoogleChatWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googledrive")
	m.Auth.Handle(urlPath, http.FileServer(http.FS(static.GoogleDriveWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googleforms")
	m.Auth.Handle(urlPath, http.FileServer(http.FS(static.GoogleFormsWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlesheets")
	m.Auth.Handle(urlPath, http.FileServer(http.FS(static.GoogleSheetsWebContent)))

	h := NewHTTPHandler(l, o, v, d)
	common.RegisterSaveHandler(m, desc, h.handleCreds)
	common.RegisterOAuthHandler(m, desc, h.handleOAuth)

	// Event webhooks (unauthenticated by definition).
	m.NoAuth.HandleFunc("POST "+calWebhookPath, h.handleCalNotification)
	m.NoAuth.HandleFunc("POST "+driveWebhookPath, h.handleDriveNotification)
	m.NoAuth.HandleFunc("POST "+formsWebhookPath, h.handleFormsNotification)
	m.NoAuth.HandleFunc("POST "+gmailWebhookPath, h.handleGmailNotification)
}
