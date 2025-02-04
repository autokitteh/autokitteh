package zoom

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"golang.org/x/oauth2"
)

var (
	clientIDName     = sdktypes.NewSymbol("client_id")
	redirectURIName  = sdktypes.NewSymbol("redirect_uri")
	clientSecretName = sdktypes.NewSymbol("client_secret")
)

var tokenResp struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func newOAuthData(t *oauth2.Token) {}
