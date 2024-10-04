package temporalclient

import (
	"time"

	"go.temporal.io/sdk/workflow"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
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

func (ac ActivityConfig) Validate() error { return nil }

// other overrides self.
func (ac ActivityConfig) With(other ActivityConfig) ActivityConfig {
	return ActivityConfig{
		StartToCloseTimeout:    kittehs.FirstNonZero(other.StartToCloseTimeout, ac.StartToCloseTimeout),
		ScheduleToCloseTimeout: kittehs.FirstNonZero(other.ScheduleToCloseTimeout, ac.ScheduleToCloseTimeout),
		HeartbeatTimeout:       kittehs.FirstNonZero(other.HeartbeatTimeout, ac.HeartbeatTimeout, defaultActivityConfig.HeartbeatTimeout),
		ScheduleToStartTimeout: kittehs.FirstNonZero(other.ScheduleToStartTimeout, ac.ScheduleToStartTimeout),
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
