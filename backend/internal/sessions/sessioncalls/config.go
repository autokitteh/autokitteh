package sessioncalls

import (
	"time"

	"go.temporal.io/sdk/worker"
)

type Config struct {
	Temporal TemporalConfig `koanf:"temporal"`
}

type TemporalConfig struct {
	ActivityHeartbeatInterval      time.Duration `koanf:"activity_heartbeat_interval"`
	ActivityHeartbeatTimeout       time.Duration `koanf:"activity_heartbeat_timeout"`
	ActivityScheduleToCloseTimeout time.Duration `koanf:"activity_schedule_to_close_timeout"`
	ActivityStartToCloseTimeout    time.Duration `koanf:"activity_start_to_close_timeout"`
	LocalScheduleToCloseTimeout    time.Duration `koanf:"local_schedule_to_close_timeout"`
	Worker                         worker.Options
}
