package common

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	AuthTypeVar = sdktypes.NewSymbol("auth_type")

	OAuthAccessTokenVar  = sdktypes.NewSymbol("oauth_access_token")
	OAuthExpiryVar       = sdktypes.NewSymbol("oauth_expiry")
	OAuthRefreshTokenVar = sdktypes.NewSymbol("oauth_refresh_token")
	OAuthTokenTypeVar    = sdktypes.NewSymbol("oauth_token_type")

	LegacyOAuthAccessTokenVar = sdktypes.NewSymbol("oauth_AccessToken")
)

// OAuthData contains OAuth 2.0 token details.
type OAuthData struct {
	AccessToken  string `var:"oauth_access_token,secret"`
	Expiry       string `var:"oauth_expiry"`
	RefreshToken string `var:"oauth_refresh_token,secret"`
	TokenType    string `var:"oauth_token_type"`
}

// ToToken converts OAuthData to an OAuth 2.0 token. If the expiry
// is missing or invalid (not RFC-3339), we use the zero time.
func (o OAuthData) ToToken() *oauth2.Token {
	expiry, err := time.Parse(time.RFC3339, o.Expiry)
	if err != nil {
		expiry = time.Time{}
	}
	return &oauth2.Token{
		AccessToken:  o.AccessToken,
		Expiry:       expiry,
		RefreshToken: o.RefreshToken,
		TokenType:    o.TokenType,
	}
}

// CheckOAuthToken returns a warning status if [OAuthAccessTokenVar] is missing in [sdktypes.Vars]; otherwise,
// it returns OK. It depends on [sdktypes.Vars] being preloaded with [OAuthAccessTokenVar], which isn't validated.
// This is reused in connection status and test functions of all integrations.
func CheckOAuthToken(vs sdktypes.Vars) (sdktypes.Status, error) {
	if vs.GetValue(OAuthAccessTokenVar) == "" {
		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
	}
	return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth 2.0"), nil
}

// CheckLegacyOAuthToken returns a warning status if [LegacyOAuthAccessTokenVar] is missing in [sdktypes.Vars]; otherwise,
// it returns OK. It depends on [sdktypes.Vars] being preloaded with [LegacyOAuthAccessTokenVar], which isn't validated.
// This is reused in connection status and test functions of all integrations.
func CheckLegacyOAuthToken(vs sdktypes.Vars) (sdktypes.Status, error) {
	if vs.GetValue(LegacyOAuthAccessTokenVar) == "" {
		return sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
	}
	return sdktypes.NewStatus(sdktypes.StatusCodeOK, "Using OAuth 2.0"), nil
}

// EncodeOAuthData encodes an OAuth 2.0 token into AutoKitteh connection variables.
func EncodeOAuthData(t *oauth2.Token) OAuthData {
	return OAuthData{
		AccessToken:  t.AccessToken,
		Expiry:       t.Expiry.Format(time.RFC3339),
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
	}
}

// FreshOAuthToken returns the OAuth token stored in the
// connection variables. If it's stale, we refresh it first.
// A token without an expiry is considered fresh forever, so
// time-limited tokens with a missing timestamp need to add it.
// Refreshed tokens are saved back to the connection variables.
func FreshOAuthToken(ctx context.Context, l *zap.Logger, o sdkservices.OAuth, v sdkservices.Vars, i sdktypes.Integration, vs sdktypes.Vars) *oauth2.Token {
	data := new(OAuthData)
	vs.Decode(data)
	t1 := data.ToToken()

	// Access token is still fresh - return it as-is.
	if t1.Valid() {
		return t1
	}

	// Otherwise, use the OAuth refresh flow.
	intg := i.UniqueName().String()
	cfg, _, err := o.Get(ctx, intg)
	if err != nil {
		l.Error("failed to get OAuth config to refresh a token",
			zap.String("integration", intg), zap.Error(err),
		)
		return t1
	}

	t2, err := cfg.TokenSource(ctx, t1).Token()
	if err != nil {
		return t1
	}

	// Special case: Salesforce access tokens are time-limited and yet
	// they don't have an expiry timestamp - so we add it on our own.
	if i.UniqueName().String() == "salesforce" && t2.Expiry.IsZero() {
		// TODO(INT-322): Reuse "accessTokenExpiration" in SFDC's OAuth handler.
		t2.Expiry = time.Now().UTC().Add(2 * time.Hour)
	}

	vsid := vs.Get(OAuthAccessTokenVar).ScopeID()
	l.Debug("refreshed OAuth token",
		zap.String("integration", intg),
		zap.String("connection_id", vsid.String()),
		zap.Time("new_expiry", t2.Expiry),
	)

	// Update the connection variables before returning the new token.
	vs = sdktypes.EncodeVars(EncodeOAuthData(t2))
	if err := v.Set(ctx, vs.WithScopeID(vsid)...); err != nil {
		l.Error("failed to save refreshed OAuth token in connection",
			zap.String("integration", intg),
			zap.String("connection_id", vsid.String()),
			zap.Error(err),
		)
	}

	return t2
}

// ReadConnectionVars returns the connection's variables, or
// an error if the connection is not initialized or accessible.
// This is reused in connection status and test functions of all integrations.
func ReadConnectionVars(ctx context.Context, vars sdkservices.Vars, cid sdktypes.ConnectionID) (sdktypes.Vars, sdktypes.Status, error) {
	if !cid.IsValid() {
		return nil, sdktypes.NewStatus(sdktypes.StatusCodeWarning, "Init required"), nil
	}

	vs, err := vars.Get(ctx, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return nil, sdktypes.InvalidStatus, err // This is abnormal.
	}

	return vs, sdktypes.InvalidStatus, nil
}

func ReadAuthType(vs sdktypes.Vars) string {
	return vs.GetValue(AuthTypeVar)
}

// SaveAuthType saves the authentication type that the user selected for a connection.
// This will be redundant if/when the only way to initialize connections is via the web UI,
// therefore we do not care if this function fails to save it as a connection variable.
func SaveAuthType(r *http.Request, vars sdkservices.Vars, vsid sdktypes.VarScopeID) string {
	authType := r.FormValue("auth_type")
	v := sdktypes.NewVar(AuthTypeVar).SetValue(authType).WithScopeID(vsid)
	_ = vars.Set(r.Context(), v)
	return authType
}
