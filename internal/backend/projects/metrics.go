package projects

import (
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.opentelemetry.io/otel/metric"
)

var projectsCreatedCounter metric.Int64Counter

func initMetrics(t *telemetry.Telemetry) {
	projectsCreatedCounter, _ = t.NewCounter("projects.created", "Created prjects counter")
}
