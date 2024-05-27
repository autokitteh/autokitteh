package authsessions

import (
	"net/http"

	"github.com/dghubble/sessions"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	Cookie     *sessions.CookieConfig
	CookieKeys string `koanf:"cookie_keys"` // pairs of hash and block keys.
}

var productionCookie = func() *sessions.CookieConfig {
	c := sessions.DefaultCookieConfig
	c.SameSite = http.SameSiteNoneMode
	return c
}()

var Configs = configset.Set[Config]{
	Default: &Config{
		Cookie: productionCookie,
	},
	Dev: &Config{
		Cookie:     sessions.DebugCookieConfig,
		CookieKeys: "0000000000000000000000000000000000000000000000000000000000000000,0000000000000000000000000000000000000000000000000000000000000000",
	},
}
