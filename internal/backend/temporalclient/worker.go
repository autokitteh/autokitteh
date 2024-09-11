package temporalclient

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type WorkerConfig struct {
	WorkflowDeadlockTimeout time.Duration `koanf:"workflow_deadlock_timeout"`
}

// other overrides self.
func (wc WorkerConfig) With(other WorkerConfig) WorkerConfig {
	return WorkerConfig{
		WorkflowDeadlockTimeout: kittehs.Choose(other.WorkflowDeadlockTimeout, wc.WorkflowDeadlockTimeout),
	}
}

func NewWorker(l *zap.Logger, client client.Client, qname string, cfg WorkerConfig) worker.Worker {
	opts := worker.Options{
		DisableRegistrationAliasing: true,
		DeadlockDetectionTimeout:    cfg.WorkflowDeadlockTimeout,
		OnFatalError:                func(err error) { l.Error(fmt.Sprintf("temporal worker: %v", err), zap.Error(err)) },
		Identity:                    qname,
	}

	return worker.New(client, qname, opts)
}
