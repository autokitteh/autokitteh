package temporalclient

import (
	"cmp"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

var defaultWorkerConfig = WorkerConfig{
	MaxConcurrentWorkflowTaskExecutionSize: 50,
	MaxConcurrentActivityExecutionSize:     50,
}

// Common way to define configuration that can be used in multiple modules,
// saving the need to repeat the same configuration in each module.
type WorkerConfig struct {
	Disable                                bool          `koanf:"disable"`
	WorkflowDeadlockTimeout                time.Duration `koanf:"workflow_deadlock_timeout"`
	MaxConcurrentWorkflowTaskExecutionSize int           `koanf:"max_concurrent_workflow_task_execution_size"`
	MaxConcurrentActivityExecutionSize     int           `koanf:"max_concurrent_activity_execution_size"`
}

// other overrides self.
func (wc WorkerConfig) With(other WorkerConfig) WorkerConfig {
	return WorkerConfig{
		WorkflowDeadlockTimeout: cmp.Or(other.WorkflowDeadlockTimeout, wc.WorkflowDeadlockTimeout),
	}
}

// NewWorker creates a new Temporal worker. If the worker is disabled, returns nil.
func NewWorker(l *zap.Logger, client client.Client, qname string, cfg WorkerConfig) worker.Worker {
	if cfg.Disable {
		l.With(zap.String("queue_name", qname)).Info(fmt.Sprintf("temporal worker for queue %q is disabled", qname))
		return nil
	}

	var interceptors []interceptor.WorkerInterceptor
	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(
		opentelemetry.TracerOptions{
			Tracer:         telemetry.T(),
			SpanContextKey: spanContextKey,
		},
	)
	if err != nil {
		l.Warn("Unable to create interceptor", zap.Error(err))
	} else {
		interceptors = append(interceptors, tracingInterceptor)
	}

	cfg = defaultWorkerConfig.With(cfg)
	opts := worker.Options{
		DisableRegistrationAliasing:            true,
		DeadlockDetectionTimeout:               cfg.WorkflowDeadlockTimeout,
		OnFatalError:                           func(err error) { l.Error(fmt.Sprintf("temporal worker: %v", err), zap.Error(err)) },
		Identity:                               fmt.Sprintf("%s__%s", qname, fixtures.ProcessID()),
		MaxConcurrentWorkflowTaskExecutionSize: cfg.MaxConcurrentWorkflowTaskExecutionSize,
		MaxConcurrentActivityExecutionSize:     cfg.MaxConcurrentActivityExecutionSize,
		Interceptors:                           interceptors,
	}

	return worker.New(client, qname, opts)
}
