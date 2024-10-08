package webtools

import "go.autokitteh.dev/autokitteh/internal/backend/config"

type Config struct {
	Enabled bool `koanf:"enabled"`
}

func (Config) Validate() error { return nil }

var Configs = config.Set[Config]{
	Default: &Config{},
	Dev:     &Config{Enabled: true},
}
