package builds

import (
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.opentelemetry.io/otel/metric"
)

var buildsCreatedCounter metric.Int64Counter

func initMetrics(t *telemetry.Telemetry) {
	buildsCreatedCounter, _ = t.NewCounter("builds.created", "Created builds counter")
}
