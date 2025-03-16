package oauth

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Config struct {
	Address string `koanf:"address"` // Prefix: webhooks
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}

type OAuth struct {
	cfg    *Config
	logger *zap.Logger
	vars   sdkservices.Vars

	BaseURL      string
	oauthConfigs map[string]oauthConfig
}

func New(c *Config, l *zap.Logger, v sdkservices.Vars) *OAuth {
	return &OAuth{cfg: c, logger: l, vars: v}
}

func (o *OAuth) Start(m *muxes.Muxes) error {
	if err := o.normalizeAddress(); err != nil {
		return fmt.Errorf("invalid server config: %w", err)
	}

	o.initConfigs()

	// TODO: Change to "/oauth" once the old OAuth service is removed.
	if m != nil {
		m.Auth.HandleFunc("GET /oauth2/start/{key}", o.startOAuthFlow)
		m.NoAuth.HandleFunc("GET /oauth2/redirect/{key}", o.exchangeCodeToToken)
	}
	return nil
}

func (o *OAuth) normalizeAddress() error {
	// Construct a URL from the address.
	a := o.cfg.Address
	if a == "" {
		a = os.Getenv("WEBHOOK_ADDRESS") // Legacy
	}
	if a == "" {
		o.BaseURL = "https://address-not-configured"
		return nil
	}
	if strings.HasPrefix(a, "http://") {
		a = strings.Replace(a, "http://", "https://", 1)
	}
	if !strings.HasPrefix(a, "https://") {
		a = "https://" + a
	}

	// Parse and normalize the URL.
	u, err := url.Parse(a)
	if err != nil {
		return fmt.Errorf("webhooks.address = %q: %w", o.cfg.Address, err)
	}
	if u.Host == "" {
		return fmt.Errorf("webhooks.address = %q: missing host", o.cfg.Address)
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

	// Set it as the AutoKitteh server's base URL for webhooks.
	o.BaseURL = u.String()
	return nil
}
