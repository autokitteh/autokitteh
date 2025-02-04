package zoom

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

func (h handler) handleSave(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Check "Content-Type" header.  //TODO: maybe this is unnacessary
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

	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")

	if clientID == "" || redirectURI == "" {
		c.AbortBadRequest("missing required fields: client_id or redirect_uri")
		return
	}

	vs := sdktypes.NewVars().
		Set(clientIDName, clientID, false).
		Set(redirectURIName, redirectURI, false).
		Set(clientSecretName, clientSecret, true)

	if err := h.saveAuthCredentials(r.Context(), c, vs); err != nil {
		l.Error("Failed to save Zoom OAuth credentials", zap.Error(err))
		c.AbortServerError("failed to save credentials")
		return
	}

	// Redirect to Zoom's OAuth authorization page.
	startZoomOAuth(w, r, clientID, redirectURI, l)
}

func startZoomOAuth(w http.ResponseWriter, r *http.Request, clientID, redirectURI string, l *zap.Logger) {
	zoomAuthURL := "https://zoom.us/oauth/authorize"

	authURL := fmt.Sprintf(
		"%s?response_type=code&client_id=%s&redirect_uri=%s",
		zoomAuthURL, clientID, redirectURI,
	)

	l.Info("Redirecting to Zoom OAuth", zap.String("auth_url", authURL))
	http.Redirect(w, r, authURL, http.StatusFound)
}

// saveAuthCredentials stores the Zoom OAuth credentials.
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
