package builds

import (
	"go.opentelemetry.io/otel/metric"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

var buildsCreatedCounter metric.Int64Counter

func initMetrics(t *telemetry.Telemetry) {
	buildsCreatedCounter, _ = t.NewCounter("builds.created", "Created builds counter")
}
