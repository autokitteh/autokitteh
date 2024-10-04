package dispatcher

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Config struct {
	Worker   temporalclient.WorkerConfig   `koanf:"worker"`
	Workflow temporalclient.WorkflowConfig `koanf:"workflow"`
	Activity temporalclient.ActivityConfig `koanf:"activity"`
}

func (c Config) Validate() error {
	return errors.Join(
		c.Worker.Validate(),
		c.Workflow.Validate(),
		c.Activity.Validate(),
	)
}

var Configs = config.Set[Config]{
	Default: &Config{},
}
