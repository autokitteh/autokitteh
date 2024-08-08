package sessionworkflows

import (
	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.opentelemetry.io/otel/metric"
)

var (
	sessionsCreatedCounter   metric.Int64Counter
	sessionsCompletedCounter metric.Int64Counter
	sessionsErroredCounter   metric.Int64Counter
	sessionsStoppedCounter   metric.Int64Counter
)

func initMetrics(t *telemetry.Telemetry) {
	sessionsCreatedCounter, _ = t.NewCounter("sessions.created", "Created sessions counter")
	sessionsCompletedCounter, _ = t.NewCounter("sessions.completed", "Completed sessions counter")
	sessionsErroredCounter, _ = t.NewCounter("sessions.errored", "Errored sessions counter")
	sessionsStoppedCounter, _ = t.NewCounter("sessions.stopped", "Stopped sessions counter")
}
