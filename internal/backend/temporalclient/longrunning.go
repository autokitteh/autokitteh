package temporalclient

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

const minInterval = 500 * time.Millisecond

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
	var zero T

	type orErr struct {
		v   T
		err error
	}

	done := make(chan orErr, 1)

	ctx := NewWorkflowContextAsGOContext(wctx)
	ctx, cancel := context.WithCancelCause(ctx)

	go func() {
		// This can take a while to complete since it might provision resources et al,
		// but since it might call back into the workflow via the callbacks, it must
		// still run a in a workflow context rather than an activity.
		t, err := f(ctx)

		done <- orErr{t, err}
	}()

	var totalCh <-chan time.Time
	if total != 0 {
		totalCh = time.After(total)
	}

	if err := workflow.Sleep(wctx, 5*time.Second); err != nil {
		cancel(err)
		return zero, err
	}

	for {
		select {
		// case <-time.After(min(interval, minInterval)):
		// 	// don't let the temporal deadlock detector kick in.
		// 	fmt.Println("here")
		// 	if err := workflow.Sleep(wctx, time.Millisecond); err != nil {
		// 		cancel(err)
		// 		return zero, err
		// 	}
		case r := <-done:
			if r.err != nil {
				return zero, r.err
			}

			return r.v, nil
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-totalCh:
			err := fmt.Errorf("slow operation timed out")
			cancel(err)
			return zero, err
		}
	}
}
