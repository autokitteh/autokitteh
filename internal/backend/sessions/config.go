package sessions

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type ExternalStartConfig struct {
	Enabled bool `koanf:"enabled"`
}

type Config struct {
	EnableWorker  bool                    `koanf:"enable_worker"`
	Workflows     sessionworkflows.Config `koanf:"workflows"`
	Calls         sessioncalls.Config     `koanf:"calls"`
	ExternalStart ExternalStartConfig     `koanf:"external_start"`

	EnableNondurableSessions bool `koanf:"enable_nondurable_sessions"`
}

var defaultConfig = Config{
	EnableWorker: true,
	ExternalStart: ExternalStartConfig{
		Enabled: false,
	},
	Workflows: sessionworkflows.Config{
		Worker: temporalclient.WorkerConfig{
			WorkflowDeadlockTimeout: time.Second * 10, // TODO: bring down to 1s.
		},
		NextEventInActivityPollDuration: time.Millisecond * 100,
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

var Configs = configset.Set[Config]{
	Default: &defaultConfig,
	Dev: func() *Config {
		c := defaultConfig
		c.Workflows.Test = true
		c.EnableNondurableSessions = true
		return &c
	}(),
}
