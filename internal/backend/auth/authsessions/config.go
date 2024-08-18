package authsessions

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	SameSite   http.SameSite
	CookieKeys string `koanf:"cookie_keys"` // pairs of hash and block keys.
	Domain     string `koanf:"ui_domain"`
	Secure     bool
}

var Configs = configset.Set[Config]{
	Default: &Config{
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	},
	Dev: &Config{
		Secure:     false,
		SameSite:   http.SameSiteLaxMode,
		CookieKeys: "0000000000000000000000000000000000000000000000000000000000000000,0000000000000000000000000000000000000000000000000000000000000000",
		Domain:     "localhost",
	},
}
