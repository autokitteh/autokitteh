package sessioncalls

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Config struct {
	// common
	Worker temporalclient.WorkerConfig `koanf:"worker"`

	// override above
	GeneralWorker temporalclient.WorkerConfig `koanf:"general_worker"`
	UniqueWorker  temporalclient.WorkerConfig `koanf:"unique_worker"`

	Activity       temporalclient.ActivityConfig `koanf:"activity"`
	UniqueActivity temporalclient.ActivityConfig `koanf:"unique_activity"`

	ActivityHeartbeatInterval time.Duration `koanf:"activity_heartbeat_interval"`
}

func (c Config) activityConfig() temporalclient.ActivityConfig {
	return c.Activity
}

func (c Config) uniqueActivityConfig() temporalclient.ActivityConfig {
	return c.Activity.With(c.UniqueActivity)
}
