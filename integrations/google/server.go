package google

import (
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

const (
	// credsPath is the URL path for our handler to save a new autokitteh
	// credentials-based connection, after the user submits it via a web form.
	credsPath = "/google/save"

	// oauthPath is the URL path for our handler to save new OAuth-based connections.
	oauthPath = "/google/oauth"

	// formsWebhookPath is the URL path to receive incoming Google Forms push notifications.
	formsWebhookPath = "/googleforms/notif"
	// gmailWebhookPath is the URL path to receive incoming Gmail push notifications.
	gmailWebhookPath = "/gmail/notif"
)

func Start(l *zap.Logger, muxNoAuth *http.ServeMux, muxAuth *http.ServeMux, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	// Note: there is a need to set some variable before calling `finalize' in `handleOauth'.
	// This could be possible only if there is authenticated user present in context (otherwise DB layer will reject the operations).
	// Therefore, oauthPath is routed via muxAuth, which will pass through auth middleware and extract authenticated user
	// from the cookie to update the context.

	uiPath := "GET " + desc.ConnectionURL().Path + "/"

	// New connection UIs + handlers.
	muxNoAuth.Handle(uiPath, http.FileServer(http.FS(static.GoogleWebContent)))

	urlPath := strings.ReplaceAll(uiPath, "google", "gmail")
	muxNoAuth.Handle(urlPath, http.FileServer(http.FS(static.GmailWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlecalendar")
	muxNoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleCalendarWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlechat")
	muxNoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleChatWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googledrive")
	muxNoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleDriveWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googleforms")
	muxNoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleFormsWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlesheets")
	muxNoAuth.Handle(urlPath, http.FileServer(http.FS(static.GoogleSheetsWebContent)))

	h := NewHTTPHandler(l, o, v, d)
	muxAuth.HandleFunc("GET "+oauthPath, h.handleOAuth)
	muxNoAuth.HandleFunc("POST "+credsPath, h.handleCreds)

	// Event webhooks.
	muxNoAuth.HandleFunc("POST "+formsWebhookPath, h.handleFormsNotification)
	muxNoAuth.HandleFunc("POST "+gmailWebhookPath, h.handleGmailNotification)
}
