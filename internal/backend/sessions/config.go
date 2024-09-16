package sessions

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Config struct {
	EnableWorker bool                    `koanf:"enable_worker"`
	Workflows    sessionworkflows.Config `koanf:"workflows"`
	Calls        sessioncalls.Config     `koanf:"calls"`
}

var defaultConfig = Config{
	EnableWorker: true,
	Workflows: sessionworkflows.Config{
		Worker: temporalclient.WorkerConfig{
			WorkflowDeadlockTimeout: time.Second * 10, // TODO: bring down to 1s.
		},
	},
	Calls: sessioncalls.Config{
		// Not sure 15s (taken from "Dev" - see below) is a good default,
		// but without a non-zero value AK panics when starting workflows.
		ActivityHeartbeatInterval: time.Second * 15,
		UniqueActivity: temporalclient.ActivityConfig{
			ScheduleToStartTimeout: time.Second * 5,
		},
	},
}

var Configs = configset.Set[Config]{
	Default: &defaultConfig,
	Dev: func() *Config {
		c := defaultConfig
		// Moved to "defaultConfig" - see above:
		// c.Calls.ActivityHeartbeatInterval = time.Second * 15
		c.Calls.Activity.HeartbeatTimeout = time.Minute
		c.Workflows.OSModule = true
		c.Workflows.Test = true
		return &c
	}(),
}
