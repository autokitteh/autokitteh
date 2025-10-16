package oauth

import (
	"fmt"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

const defaultPublicBackendBaseURL = "https://address-not-configured"

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
	// Set the AutoKitteh server's base URL for webhooks.
	o.BaseURL = o.cfg.Address
	if o.BaseURL == "" {
		o.BaseURL = fixtures.ServiceAddress()
	}

	var err error
	o.BaseURL, err = normalizeAddress(o.BaseURL, defaultPublicBackendBaseURL)
	if err != nil {
		return fmt.Errorf("invalid server webhooks config: %w", err)
	}

	// Initialize all the default OAuth 2.0 configurations.
	o.initConfigs()

	// Register the server's general-purpose OAuth 2.0 webhooks.
	if m != nil {
		m.Auth.HandleFunc("GET /oauth/start/{integration}", o.startOAuthFlow)
		m.NoAuth.HandleFunc("GET /oauth/redirect/{integration}", o.exchangeCodeToToken)
	}

	return nil
}

func normalizeAddress(addr, defaultAddr string) (string, error) {
	// Construct a URL from the address.
	if addr == "" {
		return defaultAddr, nil
	}
	if strings.HasPrefix(addr, "http://") {
		addr = strings.Replace(addr, "http://", "https://", 1)
	}
	if !strings.HasPrefix(addr, "https://") {
		addr = "https://" + addr
	}

	// Parse and normalize the URL.
	u, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("bad address %q: %w", addr, err)
	}
	if u.Host == "" {
		return "", fmt.Errorf("missing host in address %q", addr)
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}
