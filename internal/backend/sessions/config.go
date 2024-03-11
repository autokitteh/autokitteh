package sessions

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessioncalls"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionworkflows"
)

type Config struct {
	Workflows sessionworkflows.Config `koanf:"workflows"`
	Calls     sessioncalls.Config     `koanf:"calls"`
	Debug     bool                    `koanf:"debug"`
}

var defaultConfig = Config{
	Workflows: sessionworkflows.Config{
		Temporal: sessionworkflows.TemporalConfig{
			WorkflowTaskTimeout:         time.Hour,
			LocalScheduleToCloseTimeout: time.Second * 5,
		},
	},
	Calls: sessioncalls.Config{
		Temporal: sessioncalls.TemporalConfig{
			ActivityHeartbeatInterval:      time.Second * 1,
			ActivityHeartbeatTimeout:       time.Second * 3,
			ActivityScheduleToCloseTimeout: time.Hour,
			ActivityStartToCloseTimeout:    time.Minute * 30,
			LocalScheduleToCloseTimeout:    time.Second * 5,
		},
	},
}

var Configs = configset.Set[Config]{
	Default: &defaultConfig,
	Dev: func() *Config {
		c := defaultConfig
		c.Calls.Temporal.ActivityHeartbeatTimeout = time.Minute * 3
		c.Debug = true
		return &c
	}(),
}
