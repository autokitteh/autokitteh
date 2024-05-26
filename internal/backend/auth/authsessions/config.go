package authsessions

import (
	"net/http"

	"github.com/dghubble/sessions"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	Cookie     *sessions.CookieConfig `koanf:"cookie"`
	CookieKeys string                 `koanf:"cookie_keys"` // pairs of hash and block keys.
	AllowedCorsCookie bool 	         `koanf:"allowed_cors_cookie"`
}

func (c Config) cookieConfig() *sessions.CookieConfig {
	if c.Cookie == nil {
		return sessions.DefaultCookieConfig
	}

	if (c.AllowedCorsCookie){
		c.Cookie.Secure = true
		c.Cookie.SameSite = http.SameSiteNoneMode
	}
	
	return c.Cookie
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		Cookie:     sessions.DebugCookieConfig,
		CookieKeys: "0000000000000000000000000000000000000000000000000000000000000000,0000000000000000000000000000000000000000000000000000000000000000",
	},
}
