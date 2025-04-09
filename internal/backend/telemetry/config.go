package telemetry

import "go.autokitteh.dev/autokitteh/internal/backend/configset"

type Config struct {
	Enabled     bool   `koanf:"enabled"`
	ServiceName string `koanf:"service_name"`
	Endpoint    string `koanf:"endpoint"`
	Tracing     bool   `koanf:"tracing"`
}

var Configs = configset.Set[Config]{
	Default: &Config{Enabled: true, Tracing: true, ServiceName: "ak", Endpoint: "localhost:4318"},
	Dev:     &Config{Enabled: false, Tracing: true, ServiceName: "ak", Endpoint: "localhost:4318"},
}
