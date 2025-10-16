package telemetry

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

type Config struct {
	Enabled         bool    `koanf:"enabled"`
	ServiceName     string  `koanf:"service_name"`
	Endpoint        string  `koanf:"endpoint"`
	Tracing         bool    `koanf:"tracing"`
	TracingFraction float64 `koanf:"tracing_fraction"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		Enabled:         true,
		Tracing:         true,
		ServiceName:     "ak",
		Endpoint:        "localhost:4318",
		TracingFraction: 0,
	},
	Dev: &Config{
		Enabled:         false,
		Tracing:         false,
		ServiceName:     "ak",
		Endpoint:        "localhost:4318",
		TracingFraction: 1,
	},
}
