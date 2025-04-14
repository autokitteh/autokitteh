package temporalclient

import (
	"cmp"
	"time"

	"go.temporal.io/sdk/workflow"
)

var defaultActivityConfig = ActivityConfig{
	StartToCloseTimeout: 10 * time.Minute,
}

// Common way to define configuration that can be used in multiple modules,
// saving the need to repeat the same configuration in each module.
type ActivityConfig struct {
	ScheduleToCloseTimeout time.Duration `koanf:"schedule_to_close_timeout"`
	StartToCloseTimeout    time.Duration `koanf:"start_to_close_timeout"`
	HeartbeatTimeout       time.Duration `koanf:"heartbeat_timeout"`
	ScheduleToStartTimeout time.Duration `koanf:"schedule_to_start_timeout"`
}

// other overrides self.
func (ac ActivityConfig) With(other ActivityConfig) ActivityConfig {
	return ActivityConfig{
		StartToCloseTimeout:    cmp.Or(other.StartToCloseTimeout, ac.StartToCloseTimeout),
		ScheduleToCloseTimeout: cmp.Or(other.ScheduleToCloseTimeout, ac.ScheduleToCloseTimeout),
		HeartbeatTimeout:       cmp.Or(other.HeartbeatTimeout, ac.HeartbeatTimeout, defaultActivityConfig.HeartbeatTimeout),
		ScheduleToStartTimeout: cmp.Or(other.ScheduleToStartTimeout, ac.ScheduleToStartTimeout),
	}
}

func (ac ActivityConfig) ToOptions(qname string) workflow.ActivityOptions {
	ac = defaultActivityConfig.With(ac)
	return workflow.ActivityOptions{
		TaskQueue:              qname,
		ScheduleToCloseTimeout: ac.ScheduleToCloseTimeout,
		StartToCloseTimeout:    ac.StartToCloseTimeout,
		HeartbeatTimeout:       ac.HeartbeatTimeout,
		ScheduleToStartTimeout: ac.ScheduleToStartTimeout,
	}
}

func WithActivityOptions(wctx workflow.Context, qname string, ac ActivityConfig) workflow.Context {
	return workflow.WithActivityOptions(wctx, ac.ToOptions(qname))
}
