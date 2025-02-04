package height

import (
	"time"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	authTypeVar = sdktypes.NewSymbol("auth_type")
)

// oauthData contains OAuth 2.0 token details.
type oauthData struct {
	AccessToken  string `var:"oauth_access_token,secret"`
	Expiry       string `var:"oauth_expiry"`
	RefreshToken string `var:"oauth_refresh_token,secret"`
	TokenType    string `var:"oauth_token_type"`
}

func newOAuthData(t *oauth2.Token) oauthData {
	return oauthData{
		AccessToken:  t.AccessToken,
		Expiry:       t.Expiry.Format(time.RFC3339),
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
	}
}

// privateOAuth contains the user-provided details of
// a private Height OAuth 2.0 app.
type privateOAuth struct {
	ClientID      string `var:"private_client_id"`
	ClientSecret  string `var:"private_client_secret,secret"`
}
