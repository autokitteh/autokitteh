package externalclient

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	Enabled          bool   `koanf:"enabled"`
	ExternalEndpoint string `koanf:"endpoint"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		Enabled: false,
	},
}
