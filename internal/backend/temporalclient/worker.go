package temporalclient

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var defaultWorkerConfig = WorkerConfig{}

// Common way to define configuration that can be used in multiple modules,
// saving the need to repeat the same configuration in each module.
type WorkerConfig struct {
	Disable                 bool          `koanf:"disable"`
	WorkflowDeadlockTimeout time.Duration `koanf:"workflow_deadlock_timeout"`
}

func (wc WorkerConfig) Validate() error {
	return nil
}

// other overrides self.
func (wc WorkerConfig) With(other WorkerConfig) WorkerConfig {
	return WorkerConfig{
		WorkflowDeadlockTimeout: kittehs.FirstNonZero(other.WorkflowDeadlockTimeout, wc.WorkflowDeadlockTimeout),
	}
}

// NewWorker creates a new Temporal worker. If the worker is disabled, returns nil.
func NewWorker(l *zap.Logger, client client.Client, qname string, cfg WorkerConfig) worker.Worker {
	if cfg.Disable {
		l.With(zap.String("queue_name", qname)).Info(fmt.Sprintf("temporal worker for queue %q is disabled", qname))
		return nil
	}

	cfg = defaultWorkerConfig.With(cfg)
	opts := worker.Options{
		DisableRegistrationAliasing: true,
		DeadlockDetectionTimeout:    cfg.WorkflowDeadlockTimeout,
		OnFatalError:                func(err error) { l.Error(fmt.Sprintf("temporal worker: %v", err), zap.Error(err)) },
		Identity:                    qname,
	}

	return worker.New(client, qname, opts)
}
