package sessionworkflows

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.opentelemetry.io/otel/metric"
)

var sessionsCounter metric.Int64UpDownCounter

func initMetrics(t *telemetry.Telemetry) {
	sessionsCounter = t.NewOtelUpDownCounter("sessions.gaude", "The session counter")
}

func updateSessionCounter(sessionID string, sessionState sdktypes.SessionStateType) {
	labels := telemetry.Labels{"session_id": sessionID}
	val := int64(-1)

	switch sessionState {
	case sdktypes.SessionStateTypeCreated:
		labels["state"] = "created"
		val = 1
	case sdktypes.SessionStateTypeCompleted:
		labels["state"] = "completed"
	case sdktypes.SessionStateTypeStopped:
		labels["state"] = "stopped"
	case sdktypes.SessionStateTypeError:
		labels["state"] = "errored"
	}
	fmt.Println("update counter", sessionState, labels)
	sessionsCounter.Add(context.Background(), val, telemetry.WithLabels(labels))
}
