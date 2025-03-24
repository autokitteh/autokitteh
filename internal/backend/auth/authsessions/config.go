package authsessions

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	SameSite          http.SameSite
	Domain            string `koanf:"ui_domain"`
	Secure            bool   `koanf:"secure_cookie"`
	ExpirationMinutes int    `koanf:"expiration_minutes"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		SameSite:          http.SameSiteNoneMode,
		Secure:            true,
		ExpirationMinutes: 60 * 24 * 14, // 14 days
	},
	Dev: &Config{
		Secure:            false,
		SameSite:          http.SameSiteLaxMode,
		ExpirationMinutes: 60 * 24 * 14, // 14 days
	},
}
