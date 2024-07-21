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

func Start(l *zap.Logger, mux *http.ServeMux, v sdkservices.Vars, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	uiPath := "GET " + desc.ConnectionURL().Path + "/"

	// New connection UIs + handlers.
	mux.Handle(uiPath, http.FileServer(http.FS(static.GoogleWebContent)))

	urlPath := strings.ReplaceAll(uiPath, "google", "gmail")
	mux.Handle(urlPath, http.FileServer(http.FS(static.GmailWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlecalendar")
	mux.Handle(urlPath, http.FileServer(http.FS(static.GoogleCalendarWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlechat")
	mux.Handle(urlPath, http.FileServer(http.FS(static.GoogleChatWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googledrive")
	mux.Handle(urlPath, http.FileServer(http.FS(static.GoogleDriveWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googleforms")
	mux.Handle(urlPath, http.FileServer(http.FS(static.GoogleFormsWebContent)))

	urlPath = strings.ReplaceAll(uiPath, "google", "googlesheets")
	mux.Handle(urlPath, http.FileServer(http.FS(static.GoogleSheetsWebContent)))

	h := NewHTTPHandler(l, o, v, d)
	mux.HandleFunc("GET "+oauthPath, h.handleOAuth)
	mux.HandleFunc("POST "+credsPath, h.handleCreds)

	// Event webhooks.
	mux.HandleFunc("POST "+formsWebhookPath, h.handleFormsNotif)
	mux.HandleFunc("POST "+gmailWebhookPath, h.handleGmailNotif)
}
