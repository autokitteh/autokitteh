package deployments

import (
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.opentelemetry.io/otel/metric"
)

var (
	deploymentsActiveCounter   metric.Int64UpDownCounter
	deploymentsDrainingCounter metric.Int64UpDownCounter
	deploymentsInactiveCounter metric.Int64UpDownCounter
)

func initMetrics(t *telemetry.Telemetry) {
	deploymentsActiveCounter, _ = t.NewUpDownCounter("deployments.activated", "Activated deployments counter")
	deploymentsDrainingCounter, _ = t.NewUpDownCounter("deployments.drained", "Drained deployments counter")
	deploymentsDrainingCounter, _ = t.NewUpDownCounter("deployments.deactivated", "Deactivated deployments counter")
}
