package oauth

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type oauthConfig struct {
	ClientID     string `koanf:"client_id"`
	ClientSecret string `koanf:"client_secret"`
}

func (c oauthConfig) Enabled() bool {
	return c != oauthConfig{}
}

type githubConfig struct {
	oauthConfig

	EnterpriseURL string `koanf:"enterprise_url"`
	AppName       string `koanf:"app_name"`
}

type Config struct {
	WebhookAddress string `koanf:"webhook_address"`

	Github githubConfig `koanf:"github"`
	Google oauthConfig  `koanf:"google"`
	Slack  oauthConfig  `koanf:"slack"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}
