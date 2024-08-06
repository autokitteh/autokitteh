package telemetry

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var Metrics = metrics{
	Sessions: UpDownCounter{labels: labels{allowed: []string{"session_id", "state"}}},
}

type metrics struct {
	Sessions UpDownCounter
}

func UpdateSessionCounter(sessionID string, sessionState sdktypes.SessionStateType) {
	attrs := map[string]string{"session_id": sessionID}
	val := int64(-1)

	switch sessionState {
	case sdktypes.SessionStateTypeCreated:
		attrs["state"] = "created"
		val = 1
	case sdktypes.SessionStateTypeCompleted:
		attrs["state"] = "completed"
	case sdktypes.SessionStateTypeStopped:
		attrs["state"] = "stopped"
	case sdktypes.SessionStateTypeError:
		attrs["state"] = "errored"
	}
	Metrics.Sessions.Add(val, attrs)
}
