package github

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/google/go-github/v60/github"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/integrations/github/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// HandlePAT saves a new autokitteh connection with a user-submitted token.
func (h handler) handlePAT(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check the "Content-Type" header.
	if common.PostWithoutFormContentType(r) {
		ct := r.Header.Get(common.HeaderContentType)
		l.Warn("save connection: unexpected POST content type", zap.String("content_type", ct))
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("form parsing error")
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
		c.AbortBadRequest("unusable PAT error: " + err.Error())
		return
	}

	if user == nil || user.Login == nil {
		l.Warn("Unexpected response from GitHub API", zap.Any("user", user))
		c.AbortBadRequest("unexpected response from GitHub API")
		return
	}

	userJSON, _ := json.Marshal(user)
	_, patKey := filepath.Split(webhook)

	if patKey == "" {
		l.Warn("Invalid webhook URL")
		c.AbortBadRequest("invalid webhook URL")
		return
	}

	c.Finalize(sdktypes.NewVars().
		Set(vars.AuthType, "pat", false).
		Set(vars.PAT, pat, true).
		Set(vars.PATKey, patKey, false).
		Set(vars.PATSecret, secret, true).
		Set(vars.PATUser, string(userJSON), false))
}
