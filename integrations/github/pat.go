package github

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	// patPath is the URL path for our webhook to save a new autokitteh
	// PAT-based connection, after the user submits it via a web form.
	patPath = "/github/save"

	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

// HandlePAT saves a new autokitteh connection with a user-submitted token.
func (h handler) handlePAT(w http.ResponseWriter, r *http.Request) {
	l := h.logger.With(zap.String("urlPath", r.URL.Path))

	// Check the "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse inbound HTTP request", zap.Error(err))
		redirectToErrorPage(w, r, "form parsing error: "+err.Error())
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
		l.Warn("Unusable GitHub PAT", zap.Error(err))
		redirectToErrorPage(w, r, "unusable PAT error: "+err.Error())
		return
	}

	if user == nil || user.Login == nil {
		l.Warn("Unexpected response from GitHub API", zap.Any("user", user))
		redirectToErrorPage(w, r, "unexpected response from GitHub API")
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
