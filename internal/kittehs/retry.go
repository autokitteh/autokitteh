package kittehs

import (
	"context"
	"time"
)

type RetryPolicy struct {
	MaxAttempts        int // 0 unlimited. 1 means no retries.
	Interval           time.Duration
	BackoffCoefficient float64       // 0 means no backoff, same as 1.
	MaxInterval        time.Duration // 0 means no max interval.

	IsRetriable func(error) bool
}

func (r RetryPolicy) next(curr time.Duration, attempt int) (next time.Duration, done bool) {
	if attempt == 0 {
		return 0, false
	}

	if done = r.MaxAttempts != 0 && attempt >= r.MaxAttempts; done {
		return
	}

	if attempt == 1 {
		return r.Interval, false
	}

	next = curr

	if r.BackoffCoefficient > 0 {
		next = time.Duration(r.BackoffCoefficient * float64(curr))
	}

	if r.MaxInterval > 0 && next > r.MaxInterval {
		next = r.MaxInterval
	}

	return
}

func (r RetryPolicy) Execute(ctx context.Context, f func(attempt int) error) error {
	var ival time.Duration

	for attempt := 1; ; attempt++ {
		err := f(attempt - 1)
		if err == nil {
			return nil
		}

		if r.IsRetriable != nil && !r.IsRetriable(err) {
			return err
		}

		var done bool
		if ival, done = r.next(ival, attempt); done {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(ival):
			// nop
		}
	}
}
