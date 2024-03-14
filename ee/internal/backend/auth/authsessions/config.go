package authsessions

import (
	"github.com/dghubble/sessions"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	UnsafeCookie *sessions.CookieConfig `koanf:"cookie"`
	CookieKeys   []string               `koanf:"sessio_cookie_keys"` // pairs of sigining and encryption keys.
}

func (c Config) cookieConfig() *sessions.CookieConfig {
	if c.UnsafeCookie == nil {
		return sessions.DefaultCookieConfig
	}

	return c.UnsafeCookie
}

var Configs = configset.Set[Config]{
	Dev: &Config{
		UnsafeCookie: sessions.DebugCookieConfig,
		CookieKeys: []string{
			"0000000000000000000000000000000000000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000000",
		},
	},
}
