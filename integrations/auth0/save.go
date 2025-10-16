package auth0

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleCreds acts as a passthrough for the OAuth connection mode,
// to save OAuth details (custom client ID & secret, Auth0 domain).
func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
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

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	auth0Domain := r.FormValue("auth0_domain")

	if clientID == "" || clientSecret == "" || auth0Domain == "" {
		c.AbortBadRequest("missing required fields")
		return
	}

	vs := sdktypes.NewVars().
		Set(ClientIDVar, clientID, false).
		Set(ClientSecretVar, clientSecret, true).
		Set(DomainVar, auth0Domain, false)

	if err := h.saveAuthCredentials(r.Context(), c, vs); err != nil {
		l.Error("Failed to save Auth0 credentials", zap.Error(err))
		c.AbortServerError("failed to save credentials")
		return
	}

	// Redirect to AutoKitteh's OAuth starting point.
	redirectURL := fmt.Sprintf("/oauth/start/auth0?cid=%s&origin=%s", c.ConnectionID, c.Origin)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (h handler) saveAuthCredentials(ctx context.Context, c sdkintegrations.ConnectionInit, vs sdktypes.Vars) error {
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %w", err)
	}

	vsl := make([]sdktypes.Var, 0, len(vs))
	for _, v := range vs {
		vsl = append(vsl, v.WithScopeID(sdktypes.NewVarScopeID(cid)))
	}

	if err := h.vars.Set(ctx, vsl...); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}
	return nil
}
