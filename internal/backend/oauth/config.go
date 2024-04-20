package oauth

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type StateConfig struct {
	Prefix string        `koanf:"prefix"`
	TTL    time.Duration `koanf:"ttl"`
}

type Config struct {
	State StateConfig `koanf:"state"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		State: StateConfig{
			TTL: 5 * time.Minute,
		},
	},
}
