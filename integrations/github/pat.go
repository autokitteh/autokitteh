package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// patPath is the URL path for our webhook to save a new autokitteh
	// PAT-based connection, after the user submits it via a web form.
	patPath = "/github/save_pat"

	HeaderContentType = "Content-Type"
	ContentTypeForm   = "application/x-www-form-urlencoded"
)

// HandlePAT saves a new autokitteh connection with a user-submitted token.
func (h handler) handlePAT(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check "Content-Type" header.
	ct := r.Header.Get(HeaderContentType)
	if ct != ContentTypeForm {
		l.Warn("Unexpected header value",
			zap.String("header", HeaderContentType),
			zap.String("got", ct),
			zap.String("want", ContentTypeForm),
		)
		e := fmt.Sprintf("Unexpected Content-Type header: %q", ct)
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse inbound HTTP request",
			zap.Error(err),
		)
		e := "Form parsing error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	pat := r.Form.Get("pat")
	webhook := r.Form.Get("webhook")
	secret := r.Form.Get("secret")

	// Test the PAT's usability and get authoritative metadata details.
	ctx := r.Context()
	client := github.NewTokenClient(ctx, pat)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		l.Warn("Unusable Personal Access Token",
			zap.Error(err),
		)
		e := "Unusable PAT error: " + err.Error()
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	if user == nil || user.Login == nil {
		l.Warn("Unexpected response from GitHub API",
			zap.Any("user", user),
		)
		e := "Unexpected response from GitHub API"
		u := fmt.Sprintf("%serror.html?error=%s", uiPath, url.QueryEscape(e))
		http.Redirect(w, r, u, http.StatusFound)
		return
	}

	userJSON, _ := json.Marshal(user)

	_, patKey := filepath.Split(webhook)

	initData := sdktypes.NewVars().
		Set(vars.PAT, pat, true).
		Set(vars.PATKey, patKey, false).
		Set(vars.PATSecret, secret, true).
		Set(vars.PATUser, string(userJSON), false)

	sdkintegrations.FinalizeConnectionInit(w, r, integrationID, initData)
}
