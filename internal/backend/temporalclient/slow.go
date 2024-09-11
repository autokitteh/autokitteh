package temporalclient

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

func Slow[T any](
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

	for {
		select {
		case <-time.After(min(interval, 500*time.Millisecond)):
			// don't let the temporal deadlock detector kick in.
			if err := workflow.Sleep(wctx, time.Millisecond); err != nil {
				cancel(err)
				return zero, err
			}
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
