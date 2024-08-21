package deployments

import (
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.opentelemetry.io/otel/metric"
)

// deployment lifecycle: inactive (upon creation) -> active (upon activation) -> draining (optional) -> inactive

var (
	deploymentsActiveGauge    metric.Int64UpDownCounter
	deploymentsDrainingGauge  metric.Int64UpDownCounter
	deploymentsCreatedCounter metric.Int64Counter
)

func initMetrics(t *telemetry.Telemetry) {
	deploymentsActiveGauge, _ = t.NewUpDownCounter("deployments.active", "Active deployments gauge")
	deploymentsDrainingGauge, _ = t.NewUpDownCounter("deployments.draining", "Draining deployments gauge")
	deploymentsCreatedCounter, _ = t.NewCounter("deployments.created", "Created deployments counter")
}
