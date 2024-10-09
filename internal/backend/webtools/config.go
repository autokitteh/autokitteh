package webtools

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	Enabled bool `koanf:"enabled"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev:     &Config{Enabled: true},
}
