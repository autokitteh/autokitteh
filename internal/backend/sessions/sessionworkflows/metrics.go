package sessionworkflows

import (
	"go.opentelemetry.io/otel/metric"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

var (
	sessionsCreatedCounter   metric.Int64Counter
	sessionsCompletedCounter metric.Int64Counter

	sessionsErroredCounter       metric.Int64Counter
	sessionsProgramErrorsCounter metric.Int64Counter
	sessionsRetryErrorsCounter   metric.Int64Counter

	sessionsStoppedCounter     metric.Int64Counter
	sessionStaleReplaysCounter metric.Int64Counter

	sessionDurationHistogram        metric.Int64Histogram
	sessionInvocationDelayHistogram metric.Int64Histogram
)

func initMetrics(t *telemetry.Telemetry) {
	sessionsCreatedCounter, _ = t.NewCounter("sessions.created", "Created sessions counter")
	sessionsCompletedCounter, _ = t.NewCounter("sessions.completed", "Completed sessions counter")

	// erroredCounted excludes program errors and retry errors (worker health error leading to replay)
	sessionsErroredCounter, _ = t.NewCounter("sessions.errored", "Errored sessions counter")
	sessionsProgramErrorsCounter, _ = t.NewCounter("sessions.program_errors", "Program errors sessions counter")
	sessionsRetryErrorsCounter, _ = t.NewCounter("sessions.retry_errors", "Retry errors sessions counter")

	sessionsStoppedCounter, _ = t.NewCounter("sessions.stopped", "Stopped sessions counter")
	sessionStaleReplaysCounter, _ = t.NewCounter("sessions.stale_replays", "Stale replays sessions counter")

	sessionDurationHistogram, _ = t.NewHistogram("sessions.duration", "Session duration histogram")
	sessionInvocationDelayHistogram, _ = t.NewHistogram("sessions.invocation_delay", "Session invocation delay (time from event till session start) histogram")
}
