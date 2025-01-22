package microsoft

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
)

// handleSave saves connection variables for an AutoKitteh connection.
// This may result in a fully-initialized and usable connection, or it
// may be an intermediate step before starting a 3-legged OAuth 2.0 flow.
// This handler accepts both GET and POST requests alike. Why GET? This
// is the only option when the web UI opens a pop-up window for OAuth.
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check the "Content-Type" header in POST requests.
	contentType := r.Header.Get("Content-Type")
	expected := "application/x-www-form-urlencoded"
	if r.Method == http.MethodPost && !strings.HasPrefix(contentType, expected) {
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", contentType))
		c.AbortBadRequest("unexpected request content type")
		return
	}

	// Parse GET request's query params / POST request's body.
	if err := r.ParseForm(); err != nil {
		l.Warn("save connection: failed to parse HTTP request", zap.Error(err))
		c.AbortBadRequest("request parsing error")
		return
	}

	// Determine what to save and how to proceed.
	switch at := r.FormValue("auth_type"); at {
	// Use the AutoKitteh's server's default Microsoft 365 OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		if urlPath := oauthStartURL(r.Form, c); urlPath != "" {
			http.Redirect(w, r, urlPath, http.StatusFound)
		} else {
			l.Warn("save connection: bad OAuth redirect URL")
			c.AbortBadRequest("bad redirect URL")
		}

	// The the user-provided details of a custom Microsoft 365 OAuth 2.0 app.
	case integrations.OAuthCustom:
		// TODO: Implement.

	// Unknown/unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type", zap.String("auth_type", at))
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", at))
	}
}

// oauthStartURL constructs a relative URL to start a 3-legged
// OAuth 2.0 flow, using the AutoKitteh server's generic OAuth service.
func oauthStartURL(vs url.Values, c sdkintegrations.ConnectionInit) string {
	// Default scopes: all ("microsoft").
	path := "/oauth/start/microsoft-%s?cid=%s&origin=%s"

	// Narrow down the requested scopes to a specific Microsoft 365 product?
	scopes := vs.Get("auth_scopes")

	// Remember the AutoKitteh connection ID and request origin.
	path = fmt.Sprintf(path, scopes, c.ConnectionID, c.Origin)
	path = strings.ReplaceAll(path, "-?", "?")

	// Security check: ensure the URL is relative and does not contain suspicious characters.
	u, err := url.Parse(path)
	if err != nil || u.Hostname() != "" || strings.Contains(path, "..") || strings.Contains(path, "//") {
		return ""
	}

	return path
}
