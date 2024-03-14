package authloginhttpsvc

import (
	"github.com/dghubble/gologin/v2"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type oauth2Config struct {
	Enabled            bool                  `koanf:"enabled"`
	ClientID           string                `koanf:"client_id"`
	ClientSecret       string                `koanf:"client_secret"`
	RedirectURL        string                `koanf:"redirect_url"`
	UnsafeCookieConfig *gologin.CookieConfig `koanf:"cookie"`
}

func (c oauth2Config) cookieConfig() gologin.CookieConfig {
	if c.UnsafeCookieConfig == nil {
		return gologin.DefaultCookieConfig
	}

	return *c.UnsafeCookieConfig
}

type Config struct {
	GoogleOAuth oauth2Config `koanf:"google_oauth"`
	GithubOAuth oauth2Config `konf:"github_oauth"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		GoogleOAuth: oauth2Config{
			RedirectURL:        "http://localhost:9980/auth/google/callback",
			UnsafeCookieConfig: &gologin.DebugOnlyCookieConfig,
		},
		GithubOAuth: oauth2Config{
			RedirectURL:        "http://localhost:9980/auth/github/callback",
			UnsafeCookieConfig: &gologin.DebugOnlyCookieConfig,
		},
	},
}
