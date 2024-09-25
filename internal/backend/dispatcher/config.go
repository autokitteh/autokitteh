package dispatcher

import (
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Config struct {
	Worker   temporalclient.WorkerConfig   `koanf:"worker"`
	Workflow temporalclient.WorkflowConfig `koanf:"workflow"`
	Activity temporalclient.ActivityConfig `koanf:"activity"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}
