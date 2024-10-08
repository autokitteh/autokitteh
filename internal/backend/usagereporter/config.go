package usagereporter

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
)

type Config struct {
	Enabled  bool          `koanf:"enabled"`
	Endpoint string        `koanf:"endpoint"`
	Interval time.Duration `koadnf:"interval_seconds"`
}

func (Config) Validate() error { return nil }

var Configs = config.Set[Config]{
	Default: &Config{
		Enabled: false,
	},
	Dev: &Config{
		Enabled:  true,
		Endpoint: "http://localhost:9980",
		Interval: time.Hour * 24,
	},
}
