package sessionworkflows

import (
	"time"

	"go.temporal.io/sdk/worker"
)

type Config struct {
	Temporal TemporalConfig `koanf:"temporal"`
	Workflow WorkflowConfig `koanf:"workflow"`

	// Enable internal test functionality.
	Test bool `koanf:"test"`
}

type TemporalConfig struct {
	WorkflowTaskTimeout         time.Duration  `koanf:"workflow_task_timeout"`
	LocalScheduleToCloseTimeout time.Duration  `koanf:"local_schedule_to_close_timeout"`
	Worker                      worker.Options `koanf:"worker"`
}

type WorkflowConfig struct{}
