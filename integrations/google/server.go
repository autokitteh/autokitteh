package google

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/google/oauth"

	// savePath is the URL path for our handler to save new JSON key-based
	// connections, after users submit them via a web form.
	savePath = "/google/save"

	// calWebhookPath is the URL path to receive incoming Google Calendar push notifications.
	calWebhookPath = "/googlecalendar/notif"
	// formsWebhookPath is the URL path to receive incoming Google Forms push notifications.
	formsWebhookPath = "/googleforms/notif"
	// gmailWebhookPath is the URL path to receive incoming Gmail push notifications.
	gmailWebhookPath = "/gmail/notif"
)

// Start initializes all the HTTP handlers of all the Google integrations.
// This includes connection UIs, initialization webhooks, and event webhooks.
func Start(l *zap.Logger, muxes *muxes.Muxes, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	uiPath := "GET " + desc.ConnectionURL().Path + "/"

	// Connection UI.
	muxes.Main.NoAuth.Handle(uiPath, http.FileServer(http.FS(static.GoogleWebContent)))

	urlPath := strings.ReplaceAll(uiPath, "google", "gmail")
	muxes.Main.NoAuth.Handle(urlPath, http.FileServer(http.FS(static.GmailWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlecalendar")
	muxes.Main.NoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleCalendarWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlechat")
	muxes.Main.NoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleChatWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googledrive")
	muxes.Main.NoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleDriveWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googleforms")
	muxes.Main.NoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleFormsWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlesheets")
	muxes.Main.NoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleSheetsWebContent)))

	// Init webhooks save connection vars (via "c.Finalize" calls), so they need
	// to have an authenticated user context, so the DB layer won't reject them.
	// For this purpose, init webhooks are managed by the "auth" mux, which passes
	// through AutoKitteh's auth middleware to extract the user ID from a cookie.
	h := NewHTTPHandler(l, o, v, d)
	muxes.Main.Auth.HandleFunc("GET "+oauthPath, h.handleOAuth)
	muxes.Main.Auth.HandleFunc("GET "+savePath, h.handleCreds)  // Passthrough for OAuth.
	muxes.Main.Auth.HandleFunc("POST "+savePath, h.handleCreds) // Save JSON key.

	// Event webhooks (unauthenticated by definition).
	muxes.Main.NoAuth.HandleFunc("POST "+calWebhookPath, h.handleCalNotification)
	muxes.Main.NoAuth.HandleFunc("POST "+formsWebhookPath, h.handleFormsNotification)
	muxes.Main.NoAuth.HandleFunc("POST "+gmailWebhookPath, h.handleGmailNotification)
}
