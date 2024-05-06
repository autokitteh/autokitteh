package sdkintegrations

import (
	"net/url"

	"golang.org/x/oauth2"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type OAuthData struct {
	Token  *oauth2.Token
	Params url.Values
}

func (d OAuthData) Encode() (string, error) { return kittehs.EncodeURLData(d) }

func DecodeOAuthData(raw string) (data *OAuthData, err error) {
	err = kittehs.DecodeURLData(raw, &data)
	return
}

func GetOAuthDataFromURL(u *url.URL) (raw string, oauth *OAuthData, err error) {
	raw = u.Query().Get("oauth")
	oauth, err = DecodeOAuthData(raw)
	return
}
