package sessioncalls

import (
	"context"
	"time"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type callActivityInputs struct {
	SessionID     sdktypes.SessionID
	Seq           uint32
	Debug         bool
	Poller        sdktypes.Value
	AutoHeartbeat bool
}

type callActivityOutputs struct {
	Debug   any
	Attempt uint32

	// Ask for a manual retry. This is used when executorsForSessions are not
	// initialized in case the temporal workflow replays AFTER the activity respawns.
	// The manual retry will instruct the workflow to retry it with the correct
	// executorsForSessions is initialized.
	Retry bool
}

func (cs *calls) sessionCallActivity(ctx context.Context, params *callActivityInputs) (*callActivityOutputs, error) {
	executors := executorsForSessions[params.SessionID.String()]

	if params.AutoHeartbeat {
		var done func()
		ctx, done = BeginHeartbeat(ctx, cs.config.Temporal.ActivityHeartbeatInterval)
		defer done()
	}

	var (
		ret callActivityOutputs
		err error
	)

	ret.Debug, ret.Attempt, err = cs.executeCall(ctx, params.SessionID, params.Seq, params.Poller, executors, time.Now)

	if !params.Debug {
		// don't let temporal know about the debug data.
		ret.Debug = nil
	}

	return &ret, err
}
