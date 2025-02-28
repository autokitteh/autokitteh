package tempokitteh

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
)

type Worker interface {
	Start() error

	worker() worker.Worker
}

type WorkerConfig struct {
	temporalclient.WorkerConfig
	TaskQueueName string
}

type tkworker struct{ worker.Worker }

func (tk tkworker) worker() worker.Worker { return tk.Worker }
func (tk tkworker) Start() error          { return tk.Worker.Start() }

func NewWorker(l *zap.Logger, temporal client.Client, cfg WorkerConfig) Worker {
	return tkworker{
		temporalclient.NewWorker(
			l,
			temporal,
			cfg.TaskQueueName,
			cfg.WorkerConfig,
		),
	}
}
