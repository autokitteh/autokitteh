package dispatcher

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type ExternalDispatchingConfig struct {
	Enabled bool   `koanf:"enabled"`
	URL     string `koanf:"url"`
}

type Config struct {
	Worker              temporalclient.WorkerConfig   `koanf:"worker"`
	Workflow            temporalclient.WorkflowConfig `koanf:"workflow"`
	Activity            temporalclient.ActivityConfig `koanf:"activity"`
	ExternalDispatching ExternalDispatchingConfig     `koanf:"external_dispatching"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		ExternalDispatching: ExternalDispatchingConfig{
			Enabled: false,
		},
	},
}
