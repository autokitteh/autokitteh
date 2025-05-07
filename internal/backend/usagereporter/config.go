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
			Enabled:  true,
			Endpoint: "https://api.autokitteh.cloud/usage-report",
			Interval: time.Hour * 24,
		},
	}
)
