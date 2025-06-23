package common

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	ApiKeyVar   = sdktypes.NewSymbol("api_key")
	AuthTypeVar = sdktypes.NewSymbol("auth_type")

	OAuthAccessTokenVar  = sdktypes.NewSymbol("oauth_access_token")
	OAuthExpiryVar       = sdktypes.NewSymbol("oauth_expiry")
	OAuthRefreshTokenVar = sdktypes.NewSymbol("oauth_refresh_token")
	OAuthTokenTypeVar    = sdktypes.NewSymbol("oauth_token_type")

	LegacyOAuthAccessTokenVar = sdktypes.NewSymbol("oauth_AccessToken")

	PrivateClientIDVar     = sdktypes.NewSymbol("private_client_id")
	PrivateClientSecretVar = sdktypes.NewSymbol("private_client_secret")
)

// OAuthData contains OAuth 2.0 token details.
type OAuthData struct {
	AccessToken  string `var:"oauth_access_token,secret" json:"access_token"`
	Expiry       string `var:"oauth_expiry"`
	RefreshToken string `var:"oauth_refresh_token,secret" json:"refresh_token"`
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

// CheckLegacyOAuthToken returns a warning status if [LegacyOAuthAccessTokenVar] is
// missing in [sdktypes.Vars]; otherwise, it returns OK. It depends on [sdktypes.Vars]
// being preloaded with [LegacyOAuthAccessTokenVar], which isn't validated.
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

// ReadVarsWithStatus returns the connection's variables, or
// an error if the connection is not initialized or accessible.
// This is reused in connection status and test functions of all integrations.
func ReadVarsWithStatus(ctx context.Context, vars sdkservices.Vars, cid sdktypes.ConnectionID) (sdktypes.Vars, sdktypes.Status, error) {
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
