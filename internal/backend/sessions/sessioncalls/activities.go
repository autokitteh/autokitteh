package sessioncalls

import (
	"context"
	"errors"

	"go.temporal.io/sdk/activity"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	callActivityName                       = "session_call"
	createSessionCallActivityName          = "create_session_call"
	createSessionCallAttemptActivityName   = "create_session_call_attempt"
	completeSessionCallAttemptActivityName = "complete_session_call_attempt"
)

type callActivityInputs struct {
	SessionID sdktypes.SessionID
	CallSpec  sdktypes.SessionCallSpec
	Unique    bool
}

type callActivityOutputs struct {
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
		activity.RegisterOptions{Name: callActivityName},
	)

	cs.uniqueWorker.RegisterActivityWithOptions(
		cs.sessionCallActivity,
		activity.RegisterOptions{Name: callActivityName},
	)

	cs.generalWorker.RegisterActivityWithOptions(
		cs.createSessionCall,
		activity.RegisterOptions{Name: createSessionCallActivityName},
	)

	cs.generalWorker.RegisterActivityWithOptions(
		cs.createSessionCallAttempt,
		activity.RegisterOptions{Name: createSessionCallAttemptActivityName},
	)

	cs.generalWorker.RegisterActivityWithOptions(
		cs.completeSessionCallAttempt,
		activity.RegisterOptions{Name: completeSessionCallAttemptActivityName},
	)
}

func (cs *calls) sessionCallActivity(ctx context.Context, params *callActivityInputs) (*callActivityOutputs, error) {
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

		return &callActivityOutputs{Retry: true}, nil
	}

	ctx, done := BeginHeartbeat(ctx, cs.config.ActivityHeartbeatInterval)
	defer done()

	result, err := cs.executeCall(ctx, params.CallSpec, executors)
	if err != nil {
		return nil, err
	}

	return &callActivityOutputs{Result: result}, nil
}

func (cs *calls) createSessionCall(ctx context.Context, sid sdktypes.SessionID, data sdktypes.SessionCallSpec) error {
	err := cs.svcs.DB.CreateSessionCall(ctx, sid, data)
	if errors.Is(err, sdkerrors.ErrAlreadyExists) {
		err = nil
	}

	return err
}

func (cs *calls) createSessionCallAttempt(ctx context.Context, sid sdktypes.SessionID, seq uint32) (uint32, error) {
	attempt, err := cs.svcs.DB.StartSessionCallAttempt(ctx, sid, seq)
	if errors.Is(err, sdkerrors.ErrAlreadyExists) {
		err = nil
	}

	return attempt, err
}

func (cs *calls) completeSessionCallAttempt(ctx context.Context, sid sdktypes.SessionID, seq, attempt uint32, complete sdktypes.SessionCallAttemptComplete) error {
	err := cs.svcs.DB.CompleteSessionCallAttempt(ctx, sid, seq, attempt, complete)
	if errors.Is(err, sdkerrors.ErrAlreadyExists) {
		err = nil
	}

	return err
}
