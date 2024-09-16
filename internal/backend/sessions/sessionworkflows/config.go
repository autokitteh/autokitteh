package sessionworkflows

import (
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Config struct {
	SessionWorkflow     temporalclient.WorkflowConfig `koanf:"session_workflow"`
	TerminationWorkflow temporalclient.WorkflowConfig `koanf:"termination_workflow"`

	Activity temporalclient.ActivityConfig `koanf:"activity"`

	Worker temporalclient.WorkerConfig `koanf:"worker"`

	// Enable internal test functionality.
	OSModule bool `koanf:"os_module"`

	SlowOperationTimeout time.Duration `koanf:"slow_operation_timeout"`
}
