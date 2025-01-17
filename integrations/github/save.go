package github

import (
	"context"
	"fmt"
	"net/http"
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

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	appID := r.FormValue("app_id")
	webhookSecret := r.FormValue("webhook_secret")
	enterpriseURL := r.FormValue("enterprise_url")
	privateKey := r.FormValue("private_key")

	if clientID == "" || clientSecret == "" {
		c.AbortBadRequest("missing client ID or client secret")
		return
	}

	if err := h.saveClientIDAndSecret(r.Context(), c, clientID, clientSecret, appID, webhookSecret, enterpriseURL, privateKey); err != nil {
		l.Warn("Failed to save client ID and secret", zap.Error(err))
		c.AbortBadRequest("failed to save client ID and secret")
		return
	}

	redirectURL := fmt.Sprintf("/oauth/start/github?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h handler) saveClientIDAndSecret(ctx context.Context, c sdkintegrations.ConnectionInit, clientID, clientSecret, appID, webhookSecret, enterpriseURL, privateKey string) error {
	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	scopeID := sdktypes.NewVarScopeID(cid)
	vars := []sdktypes.Var{
		sdktypes.NewVar(vars.ClientID).SetValue(clientID).WithScopeID(scopeID),
		sdktypes.NewVar(vars.ClientSecret).SetValue(clientSecret).WithScopeID(scopeID).SetSecret(true),
		sdktypes.NewVar(vars.AppID).SetValue(appID).WithScopeID(scopeID),
		sdktypes.NewVar(vars.WebhookSecret).SetValue(webhookSecret).WithScopeID(scopeID).SetSecret(true),
		sdktypes.NewVar(vars.EnterpriseURL).SetValue(enterpriseURL).WithScopeID(scopeID),
		sdktypes.NewVar(vars.PrivateKey).SetValue(privateKey).WithScopeID(scopeID).SetSecret(true),
	}

	if err := h.vars.Set(ctx, vars...); err != nil {
		return err
	}

	return nil
}

func (h handler) isCustomOAuth(ctx context.Context, cid sdktypes.ConnectionID) bool {
	vs, err := h.vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return false
	}
	return vs.GetValueByString("client_secret") != ""
}
