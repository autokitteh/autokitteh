package usagereporter

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	Enabled  bool          `koanf:"enabled"`
	Endpoint string        `koanf:"endpoint"`
	Interval time.Duration `koadnf:"interval_seconds"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{
			Enabled: false,
		},
		Dev: &Config{
			Enabled:  true,
			Endpoint: "http://localhost:9980",
			Interval: time.Hour * 24,
		},
	}
)
