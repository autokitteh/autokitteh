package deployments

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

// deployment lifecycle: inactive (upon creation) -> active (upon activation) -> draining (optional) -> inactive

var (
	deploymentsActiveGauge    metric.Int64UpDownCounter = noop.Int64UpDownCounter{}
	deploymentsDrainingGauge  metric.Int64UpDownCounter = noop.Int64UpDownCounter{}
	deploymentsCreatedCounter metric.Int64Counter       = noop.Int64Counter{}
)

func initMetrics() {
	deploymentsActiveGauge, _ = telemetry.NewUpDownCounter("deployments.active", "Active deployments gauge")
	deploymentsDrainingGauge, _ = telemetry.NewUpDownCounter("deployments.draining", "Draining deployments gauge")
	deploymentsCreatedCounter, _ = telemetry.NewCounter("deployments.created", "Created deployments counter")
}
