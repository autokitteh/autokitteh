package google

import (
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/web/static"
)

func Start(l *zap.Logger, mux *http.ServeMux, s sdkservices.Secrets, o sdkservices.OAuth, d sdkservices.Dispatcher) {
	if !checkRequiredEnvVars(l) {
		return
	}

	// New connection UIs + handlers.
	h := NewHTTPHandler(l, s, o, "google")
	mux.Handle(uiPath, http.FileServer(http.FS(static.GoogleWebContent)))
	mux.HandleFunc(oauthPath, h.HandleOAuth)
	mux.HandleFunc(credsPath, h.HandleCreds)

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

	// TODO: Event webhooks.
}

func checkRequiredEnvVars(l *zap.Logger) bool {
	result := true
	for _, k := range []string{
		// OAuth
		"GOOGLE_CLIENT_ID",
		"GOOGLE_CLIENT_SECRET",
	} {
		if os.Getenv(k) == "" {
			l.Warn("Required environment variable is missing",
				zap.String("name", k),
			)
			result = false
		}
	}
	return result
}
