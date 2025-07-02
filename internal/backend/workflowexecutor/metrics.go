//go:build enterprise
// +build enterprise

package workflowexecutor

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type metrics struct {
	workerID                      string
	queuedWorkflowsGauge          metric.Int64UpDownCounter
	activeWorkflowsGauge          metric.Int64UpDownCounter
	retriedWorkflowRequestCounter metric.Int64Counter
	workerLoadGauge               metric.Float64Gauge
}

func newMetrics(workerID string) *metrics {

	return &metrics{
		workerID:                      workerID,
		queuedWorkflowsGauge:          kittehs.Must1(telemetry.NewUpDownCounter("workflow_executor.queued", "Queued workflows gauge")),
		activeWorkflowsGauge:          kittehs.Must1(telemetry.NewUpDownCounter("workflow_executor.active_workflows", "Active workflows gauge")),
		retriedWorkflowRequestCounter: kittehs.Must1(telemetry.NewCounter("workflow_executor.retried", "Retried workflows counter")),
		workerLoadGauge:               kittehs.Must1(telemetry.NewGauge("workflow_executor.load", "Worker load gauge")),
	}
}

func (m *metrics) IncrementQueuedWorkflows(ctx context.Context) {
	m.queuedWorkflowsGauge.Add(ctx, 1)
}
func (m *metrics) DecrementQueuedWorkflows(ctx context.Context) {
	m.queuedWorkflowsGauge.Add(ctx, -1)
}

func (m *metrics) IncrementActiveWorkflows(ctx context.Context) {
	m.activeWorkflowsGauge.Add(ctx, 1, metric.WithAttributes(attribute.String("worker_id", m.workerID)))
}
func (m *metrics) DecrementActiveWorkflows(ctx context.Context) {
	m.activeWorkflowsGauge.Add(ctx, -1, metric.WithAttributes(attribute.String("worker_id", m.workerID)))
}

func (m *metrics) IncrementRetriedWorkflowsCounter(ctx context.Context) {
	m.retriedWorkflowRequestCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("worker_id", m.workerID)))
}

func (m *metrics) SetWorkerLoad(ctx context.Context, load float64) {
	m.workerLoadGauge.Record(ctx, load, metric.WithAttributes(attribute.String("worker_id", m.workerID)))
}
