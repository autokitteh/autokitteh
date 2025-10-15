package notion

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/integrations"
	"go.autokitteh.dev/autokitteh/integrations/common"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// handleSave saves connection variables for an AutoKitteh connection.
// This may result in a fully-initialized and usable connection, or it
// may be an intermediate step before starting a 3-legged OAuth 2.0 flow.
// This handler accepts both GET and POST requests alike. Why GET? This
// is the only option when the web UI opens a pop-up window for OAuth.
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

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("save connection: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	vsid := sdktypes.NewVarScopeID(cid)
	authType := common.SaveAuthType(r, h.vars, vsid)
	l = l.With(zap.String("auth_type", authType))

	switch authType {
	// Use the AutoKitteh server's default Notion OAuth 2.0 app, i.e.
	// immediately redirect to the 3-legged OAuth 2.0 flow's starting point.
	case integrations.OAuthDefault:
		startOAuth(w, r, c, l)

	// Check and save the provided API key, no 3-legged OAuth 2.0 flow is needed.
	case integrations.APIKey:
		apiKey := r.FormValue("api_key")
		if apiKey == "" {
			l.Info("save connection: missing API key for connection " + cid.String())
			c.AbortBadRequest("missing API key")
			return
		}

		// Validate API key before saving.
		if err := validateAPIKey(r.Context(), apiKey); err != nil {
			l.Debug("save connection: invalid API key for connection "+cid.String(), zap.Error(err))
			c.AbortBadRequest("invalid API key, please try again or contact support")
			return
		}

		v := sdktypes.NewVar(common.ApiKeyVar).SetValue(apiKey).SetSecret(true)
		if err := h.vars.Set(r.Context(), v.WithScopeID(vsid)); err != nil {
			l.Error("failed to save vars for connection "+cid.String()+": "+err.Error(), zap.Error(err))
			c.AbortServerError("Internal error")
			return
		}

	// Unrecognized mode - an error.
	default:
		l.Warn("save connection: unexpected auth type")
		c.AbortBadRequest(fmt.Sprintf("unexpected auth type %q", authType))
	}
}

// startOAuth redirects the user to the AutoKitteh server's
// generic OAuth service, to start a 3-legged OAuth 2.0 flow.
func startOAuth(w http.ResponseWriter, r *http.Request, c sdkintegrations.ConnectionInit, l *zap.Logger) {
	// Security check: parameters must be alphanumeric strings,
	// to prevent path traversal attacks and other issues.
	re := regexp.MustCompile(`^\w+$`)
	if !re.MatchString(c.ConnectionID + c.Origin) {
		l.Warn("save connection: bad OAuth redirect URL")
		c.AbortBadRequest("bad redirect URL")
		return
	}

	urlPath := fmt.Sprintf("/oauth/start/notion?cid=%s&origin=%s&owner=user&response_type=code", c.ConnectionID, c.Origin)
	http.Redirect(w, r, urlPath, http.StatusFound)
}

// validateAPIKey validates the Notion API key by making a test API call.
func validateAPIKey(ctx context.Context, apiKey string) error {
	// Use the Notion API's /v1/users/me endpoint to validate the key.
	url := "https://api.notion.com/v1/users/me"

	// Create HTTP request with Notion-specific headers.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Notion-Version", "2022-06-28")

	// Send the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("API key validation failed: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}
