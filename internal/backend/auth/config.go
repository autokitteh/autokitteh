package auth

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

type Config struct {
	Enabled   bool   `koanf:"enabled"`
	ProjectID string `koanf:"project_id"`
	Provider  string `koanf:"provider"`
	LogLevel  string `koanf:"log_level"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		Enabled: false,
	},
}
