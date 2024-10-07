package authloginhttpsvc

import (
	"errors"

	"github.com/dghubble/gologin/v2"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
)

type oauth2Config struct {
	Enabled      bool                  `koanf:"enabled"`
	ClientID     string                `koanf:"client_id"`
	ClientSecret string                `koanf:"client_secret"`
	RedirectURL  string                `koanf:"redirect_url"`
	Cookie       *gologin.CookieConfig `koanf:"cookie"`
}

func (c oauth2Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.ClientID == "" {
		return errors.New("client_id is required")
	}

	if c.ClientSecret == "" {
		return errors.New("client_secret is required")
	}

	if c.RedirectURL == "" {
		return errors.New("redirect_url is required")
	}

	return nil
}

func (c oauth2Config) cookieConfig() gologin.CookieConfig {
	if c.Cookie == nil {
		return gologin.DefaultCookieConfig
	}

	return *c.Cookie
}

type descopeConfig struct {
	Enabled   bool   `koanf:"enabled"`
	ProjectID string `koanf:"project_id"`
}

func (c descopeConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.ProjectID == "" {
		return errors.New("project_id is required")
	}

	return nil
}

type Config struct {
	GoogleOAuth oauth2Config  `koanf:"google_oauth"`
	GithubOAuth oauth2Config  `konf:"github_oauth"`
	Descope     descopeConfig `koanf:"descope"`

	// Allowed login patterns, separated by commas.
	// Pattern format is either of:
	// - "" or "*" - matches any login
	// - "*@domain"  - matches any login, but only from domain
	// - otherwise - matches exact login
	AllowedLogins string `koanf:"allowed_logins"`
}

func (c Config) Validate() error {
	return errors.Join(
		c.GoogleOAuth.Validate(),
		c.GithubOAuth.Validate(),
		c.Descope.Validate(),
	)
}

var Configs = config.Set[Config]{
	Default: &Config{},
	Dev: &Config{
		GoogleOAuth: oauth2Config{
			RedirectURL: "http://localhost:9980/auth/google/callback",
			Cookie:      &gologin.DebugOnlyCookieConfig,
		},
		GithubOAuth: oauth2Config{
			RedirectURL: "http://localhost:9980/auth/github/callback",
			Cookie:      &gologin.DebugOnlyCookieConfig,
		},
	},
}
