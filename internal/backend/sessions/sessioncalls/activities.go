package sessioncalls

import (
	"context"

	"go.temporal.io/sdk/activity"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const CallActivityName = "session_call"

type CallActivityInputs struct {
	SessionID sdktypes.SessionID
	CallSpec  sdktypes.SessionCallSpec
	Unique    bool
}

type CallActivityOutputs struct {
	Result sdktypes.SessionCallAttemptResult

	// Ask for a manual retry. This is used when executorsForSessions are not
	// initialized in case the temporal workflow replays AFTER the activity respawns.
	// The manual retry will instruct the workflow to retry it with the correct
	// executorsForSessions is initialized.
	Retry bool
}

func (cs *calls) registerActivities() {
	cs.generalWorker.RegisterActivityWithOptions(
		cs.sessionCallActivity,
		activity.RegisterOptions{Name: CallActivityName},
	)

	cs.uniqueWorker.RegisterActivityWithOptions(
		cs.sessionCallActivity,
		activity.RegisterOptions{Name: CallActivityName},
	)
}

func (cs *calls) sessionCallActivity(ctx context.Context, params *CallActivityInputs) (*CallActivityOutputs, error) {
	sl := cs.l.Sugar().With("session_id", params.SessionID, "seq", params.CallSpec.Seq())

	cs.executorsForSessionsMu.RLock()
	executors := cs.executorsForSessions[params.SessionID]
	cs.executorsForSessionsMu.RUnlock()

	if params.Unique && executors == nil {
		// Unique execution is requested (see comment at calls.go), but the executors
		// are not initialized. This means that the matching workflow was not yet
		// replayed, but the activity was respawned. In this case we explicitly ask
		// the workflow to retry after it is initialized.

		sl.Warnf("unique execution requested for %v#%d, but executors are not initialized", params.SessionID, params.CallSpec.Seq())

		return &CallActivityOutputs{Retry: true}, nil
	}

	if !params.CallSpec.Function().GetFunction().HasFlag(sdktypes.DisableAutoHeartbeat) && cs.config.ActivityHeartbeatInterval > 0 {
		_, done := BeginHeartbeat(ctx, cs.config.ActivityHeartbeatInterval)
		defer done()
	}

	result, err := cs.executeCall(ctx, params.CallSpec, executors)
	if err != nil {
		return nil, temporalclient.TranslateError(err, "execute call for %v", params.SessionID)
	}

	return &CallActivityOutputs{Result: result}, nil
}
