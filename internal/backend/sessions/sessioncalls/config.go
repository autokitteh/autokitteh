package sessioncalls

import (
	"errors"
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

func (c Config) Validate() error {
	return errors.Join(
		c.Worker.Validate(),
		c.GeneralWorker.Validate(),
		c.UniqueWorker.Validate(),
		c.activityConfig().Validate(),
		c.uniqueActivityConfig().Validate(),
	)
}

func (c Config) activityConfig() temporalclient.ActivityConfig {
	return c.Activity
}

func (c Config) uniqueActivityConfig() temporalclient.ActivityConfig {
	return c.Activity.With(c.UniqueActivity)
}
