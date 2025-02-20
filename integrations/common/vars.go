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
	AuthTypeVar = sdktypes.NewSymbol("auth_type")

	OAuthAccessTokenVar  = sdktypes.NewSymbol("oauth_access_token")
	OAuthExpiryVar       = sdktypes.NewSymbol("oauth_expiry")
	OAuthRefreshTokenVar = sdktypes.NewSymbol("oauth_refresh_token")
	OAuthTokenTypeVar    = sdktypes.NewSymbol("oauth_token_type")
)

// OAuthData contains OAuth 2.0 token details.
type OAuthData struct {
	AccessToken  string `var:"oauth_access_token,secret"`
	Expiry       string `var:"oauth_expiry"`
	RefreshToken string `var:"oauth_refresh_token,secret"`
	TokenType    string `var:"oauth_token_type"`
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

// ReadConnectionVars returns the connection's variables, or
// an error if the connection is not initialized or accessible.
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

// RenameVar renames a variable in the given connection scope. It does nothing if
// the variable doesn't already exist. This is useful for non-trivial data migrations.
func RenameVar(ctx context.Context, v sdkservices.Vars, vsid sdktypes.VarScopeID, old, new sdktypes.Symbol) error {
	vs, err := v.Get(ctx, vsid, old)
	if err != nil {
		return err
	}

	o := vs.Get(old)
	if !o.IsValid() {
		return nil
	}

	n := sdktypes.NewVar(new).SetValue(o.Value()).SetSecret(o.IsSecret())
	if err := v.Set(ctx, n.WithScopeID(vsid)); err != nil {
		return err
	}

	return v.Delete(ctx, vsid, old)
}
