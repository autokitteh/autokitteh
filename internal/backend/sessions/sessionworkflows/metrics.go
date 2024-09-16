package sessionworkflows

import (
	"go.opentelemetry.io/otel/metric"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

var (
	sessionsCreatedCounter     metric.Int64Counter
	sessionsCompletedCounter   metric.Int64Counter
	sessionsErroredCounter     metric.Int64Counter
	sessionsStoppedCounter     metric.Int64Counter
	sessionStaleReplaysCounter metric.Int64Counter
	sessionDurationHistogram   metric.Int64Histogram
)

func initMetrics(t *telemetry.Telemetry) {
	sessionsCreatedCounter, _ = t.NewCounter("sessions.created", "Created sessions counter")
	sessionsCompletedCounter, _ = t.NewCounter("sessions.completed", "Completed sessions counter")
	sessionsErroredCounter, _ = t.NewCounter("sessions.errored", "Errored sessions counter")
	sessionsStoppedCounter, _ = t.NewCounter("sessions.stopped", "Stopped sessions counter")
	sessionDurationHistogram, _ = t.NewHistogram("sessions.duration", "Session duration histogram")
	sessionStaleReplaysCounter, _ = t.NewCounter("sessions.stale_replays", "Stale replays sessions counter")
}
