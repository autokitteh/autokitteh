package internalclient

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
)

type Config struct {
	InternalEndpoint string `koanf:"internal_endpoint"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		InternalEndpoint: "http://localhost:9980",
	},
}
