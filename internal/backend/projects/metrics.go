package projects

import (
	"go.opentelemetry.io/otel/metric"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

var projectsCreatedCounter metric.Int64Counter

func initMetrics() {
	projectsCreatedCounter, _ = telemetry.NewCounter("projects.created", "Created projects counter")
}
