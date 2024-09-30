package temporalclient

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
)

// const minInterval = 500 * time.Millisecond

// LongRunning allows to execute the `f`, which does not
// perform any temporal operations (ie calling temporal using `wctx`)
// and thus might trigger the temporal deadlock detector. This function
// runs `f` in a separate go routine while periodically yielding to
// temporal in the calling goroutine each `interval`. If `interval` is
// less than `minInterval`, it is set to `minInterval`.
//
// This function is useful for `f`s that we know are slow and should not
// have their result cached by temporal - they need to run every
// workflow invocation, including during replays. Example usage is resource
// allocation.
func LongRunning[T any](
	wctx workflow.Context,
	interval time.Duration,
	total time.Duration,
	f func(ctx context.Context) (T, error),
) (T, error) {

	var res T
	var err error
	var done bool = false
	workflow.Go(wctx, func(gCtx workflow.Context) {
		ctx := NewWorkflowContextAsGOContext(gCtx)
		res, err = f(ctx)
		done = true
	})

	if err := workflow.Await(wctx, func() bool {
		return done
	}); err != nil {
		return res, err
	}

	return res, err
}
