package authsessions

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
)

type Config struct {
	SameSite   http.SameSite
	CookieKeys string `koanf:"cookie_keys"` // pairs of hash and block keys.
	Domain     string `koanf:"ui_domain"`
	Secure     bool
}

func (c Config) Validate() error {
	_, err := parseCookieKeys(c.CookieKeys)
	return err
}

var Configs = config.Set[Config]{
	Default: &Config{
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	},
	Dev: &Config{
		Secure:     false,
		SameSite:   http.SameSiteLaxMode,
		CookieKeys: "0000000000000000000000000000000000000000000000000000000000000000,0000000000000000000000000000000000000000000000000000000000000000",
	},
}
