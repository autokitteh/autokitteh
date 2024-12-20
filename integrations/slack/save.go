package slack

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.autokitteh.dev/autokitteh/integrations/slack/internal/vars"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

const (
	headerContentType = "Content-Type"
	contentTypeForm   = "application/x-www-form-urlencoded"
)

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
		c.AbortBadRequest("form parsing error")
		return
	}

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	signingSecret := r.FormValue("signing_secret")

	if err := h.saveClientIDAndSecrets(r.Context(), c, clientID, clientSecret, signingSecret); err != nil {
		l.Warn("Failed to save client ID and secret", zap.Error(err))
		c.AbortBadRequest("failed to save OAuth configuration")
		return
	}

	redirectURL := fmt.Sprintf("/oauth/start/slack?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h handler) saveClientIDAndSecrets(ctx context.Context, c sdkintegrations.ConnectionInit, clientID, clientSecret, signingSecret string) error {
	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return err
	}

	scopeID := sdktypes.NewVarScopeID(cid)
	vars := []sdktypes.Var{
		sdktypes.NewVar(vars.ClientID).SetValue(clientID).WithScopeID(scopeID),
		sdktypes.NewVar(vars.ClientSecret).SetValue(clientSecret).WithScopeID(scopeID).SetSecret(true),
		sdktypes.NewVar(vars.SigningSecret).SetValue(signingSecret).WithScopeID(scopeID).SetSecret(true),
	}

	if err := h.vars.Set(ctx, vars...); err != nil {
		return err
	}

	return nil
}
