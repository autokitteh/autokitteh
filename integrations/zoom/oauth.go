package zoom

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func (h handler) handleOAuth(w http.ResponseWriter, r *http.Request) {
	c, l := sdkintegrations.NewConnectionInit(h.logger, w, r, desc)

	// Parse the GET request's query params.
	if err := r.ParseForm(); err != nil {
		l.Warn("save connection after OAuth flow: failed to parse HTTP request", zap.Error(err))
		c.AbortBadRequest("request parsing error")
		return
	}

	// Sanity check: the connection ID is valid.
	cid, err := sdktypes.StrictParseConnectionID(c.ConnectionID)
	if err != nil {
		l.Warn("save connection after OAuth flow: invalid connection ID", zap.Error(err))
		c.AbortBadRequest("invalid connection ID")
		return
	}

	// Handle OAuth errors
	if e := r.FormValue("error"); e != "" {
		l.Warn("OAuth redirection reported an error", zap.Error(errors.New(e)))
		c.AbortBadRequest(e)
		return
	}

	// Get authorization code from query params
	code := r.FormValue("code")
	if code == "" {
		l.Warn("OAuth redirection missing authorization code")
		c.AbortBadRequest("missing authorization code")
		return
	}

	vs, err := h.vars.Get(r.Context(), sdktypes.NewVarScopeID(cid))
	if err != nil {
		l.Error("Failed to get Auth0 vars", zap.Error(err))
		c.AbortServerError("unknown Auth0 domain")
		return
	}

	// Exchange authorization code for access token
	token, err := h.exchangeCodeForToken(r.Context(), vs, code)
	if err != nil {
		l.Error("failed to exchange authorization code for token", zap.Error(err))
		c.AbortServerError("failed to obtain access token")
		return
	}

	// Save the connection details
	// Save the connection details.
	if err := h.saveConnection(r.Context(), sdktypes.NewVarScopeID(cid), token); err != nil {
		l.Error("failed to save OAuth connection details", zap.Error(err))
		c.AbortServerError("failed to save connection details")
		return
	}

	// Redirect back to UI
	urlPath, err := c.FinalURL()
	if err != nil {
		l.Error("failed to construct final OAuth URL", zap.Error(err))
		c.AbortServerError("bad redirect URL")
		return
	}
	http.Redirect(w, r, urlPath, http.StatusFound)
}

func (h handler) exchangeCodeForToken(ctx context.Context, vs sdktypes.Vars, code string) (*oauth2.Token, error) {
	rURL := vs.GetValueByString("redirect_uri")
	client := vs.GetValueByString("client_id")
	secret := vs.GetValueByString("client_secret")

	if rURL == "" || client == "" || secret == "" {
		return nil, fmt.Errorf("missing required OAuth values: client_id, client_secret, or redirect_uri")
	}

	// Construct the token request to Zoom
	data := url.Values{}
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", rURL)

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://zoom.us/oauth/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("creating token request: %w", err)
	}

	// Add required headers
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", client, secret)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status: %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("parsing token response: %w", err)
	}

	// Convert to oauth2.Token
	return &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		TokenType:    tokenResp.TokenType,
		RefreshToken: tokenResp.RefreshToken,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
}

// saveConnection stores the OAuth tokens securely.
func (h handler) saveConnection(ctx context.Context, vsid sdktypes.VarScopeID, t *oauth2.Token) error {
	if t == nil {
		return errors.New("OAuth redirection missing token data")
	}

	vs := sdktypes.EncodeVars(newOAuthData(t))
	return h.vars.Set(ctx, vs.WithScopeID(vsid)...)
}
