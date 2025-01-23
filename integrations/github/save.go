package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/github/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleSave acts as a passthrough for the OAuth connection mode,
// to save OAuth details (custom client ID & secrets).
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.
	contentType := r.Header.Get(headerContentType)
	if !strings.HasPrefix(contentType, contentTypeForm) {
		c.AbortBadRequest("unexpected content type")
		return
	}

	// Read and parse POST request body.
	if err := r.ParseForm(); err != nil {
		l.Warn("Failed to parse incoming HTTP request", zap.Error(err))
		c.AbortBadRequest("failed to parse form data")
		return
	}

	if r.Form.Get("client_id") == "" || r.Form.Get("client_secret") == "" {
		c.AbortBadRequest("missing client ID or client secret")
		return
	}

	if err := h.saveClientIDAndSecret(r.Context(), c, r.Form); err != nil {
		l.Warn("Failed to save client ID and secret", zap.Error(err))
		c.AbortBadRequest("failed to save client ID and secret")
		return
	}

	redirectURL := fmt.Sprintf("/oauth/start/github?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h handler) saveClientIDAndSecret(ctx context.Context, c sdkintegrations.ConnectionInit, form url.Values) error {
	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	scopeID := sdktypes.NewVarScopeID(cid)
	vars := []sdktypes.Var{
		sdktypes.NewVar(vars.ClientID).SetValue(form.Get("client_id")).WithScopeID(scopeID),
		sdktypes.NewVar(vars.ClientSecret).SetValue(form.Get("client_secret")).WithScopeID(scopeID).SetSecret(true),
		sdktypes.NewVar(vars.AppID).SetValue(form.Get("app_id")).WithScopeID(scopeID),
		sdktypes.NewVar(vars.WebhookSecret).SetValue(form.Get("webhook_secret")).WithScopeID(scopeID).SetSecret(true),
		sdktypes.NewVar(vars.EnterpriseURL).SetValue(form.Get("enterprise_url")).WithScopeID(scopeID),
		sdktypes.NewVar(vars.PrivateKey).SetValue(form.Get("private_key")).WithScopeID(scopeID).SetSecret(true),
	}

	return h.vars.Set(ctx, vars...)
}
