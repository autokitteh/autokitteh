package sdkintegrations

import (
	"net/url"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type OAuthData struct {
	Token  *oauth2.Token
	Params url.Values
}

func (d OAuthData) Encode() (string, error) { return kittehs.EncodeURLData(d) }

func (d OAuthData) ToVars() sdktypes.Vars {
	data := struct {
		AccessToken  string `var:"secret"`
		TokenType    string
		RefreshToken string `var:"secret"`
		Expiry       string
	}{
		AccessToken:  d.Token.AccessToken,
		TokenType:    d.Token.TokenType,
		RefreshToken: d.Token.RefreshToken,
		Expiry:       d.Token.Expiry.String(),
	}

	return sdktypes.EncodeVars(data).WithPrefix("oauth_")
}

func DecodeOAuthData(raw string) (data *OAuthData, err error) {
	err = kittehs.DecodeURLData(raw, &data)
	return
}

func GetOAuthDataFromURL(u *url.URL) (raw string, oauth *OAuthData, err error) {
	raw = u.Query().Get("oauth")
	oauth, err = DecodeOAuthData(raw)
	return
}
