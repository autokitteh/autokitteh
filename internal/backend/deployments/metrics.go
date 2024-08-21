package deployments

import (
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.opentelemetry.io/otel/metric"
)

// deployment lifecycle: inactive (upon creation) -> active (upon activation) -> draining (optional) -> inactive

var (
	deploymentsActiveCounter   metric.Int64UpDownCounter
	deploymentsDrainingCounter metric.Int64UpDownCounter
	deploymentsCreatedCounter  metric.Int64Counter
)

func initMetrics(t *telemetry.Telemetry) {
	deploymentsActiveCounter, _ = t.NewUpDownCounter("deployments.activated", "Activated deployments counter")
	deploymentsDrainingCounter, _ = t.NewUpDownCounter("deployments.drained", "Drained deployments counter")
	deploymentsCreatedCounter, _ = t.NewCounter("deployments.created", "Created deployments counter")
}
