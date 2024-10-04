package sessions

import (
	"errors"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/config"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Config struct {
	EnableWorker bool                    `koanf:"enable_worker"`
	Workflows    sessionworkflows.Config `koanf:"workflows"`
	Calls        sessioncalls.Config     `koanf:"calls"`
}

func (c Config) Validate() error {
	if !c.EnableWorker {
		return nil
	}

	return errors.Join(
		c.Workflows.Validate(),
		c.Calls.Validate(),
	)
}

var defaultConfig = Config{
	EnableWorker: true,
	Workflows: sessionworkflows.Config{
		Worker: temporalclient.WorkerConfig{
			WorkflowDeadlockTimeout: time.Second * 10, // TODO: bring down to 1s.
		},
	},
	Calls: sessioncalls.Config{
		ActivityHeartbeatInterval: time.Second * 5,
		Activity: temporalclient.ActivityConfig{
			HeartbeatTimeout: time.Second * 20,
		},
		UniqueActivity: temporalclient.ActivityConfig{
			ScheduleToStartTimeout: time.Second * 5,
		},
	},
}

var Configs = config.Set[Config]{
	Default: &defaultConfig,
	Dev: func() *Config {
		c := defaultConfig
		c.Workflows.OSModule = true
		c.Workflows.Test = true
		return &c
	}(),
}
