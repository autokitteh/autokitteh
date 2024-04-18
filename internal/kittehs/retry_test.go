package kittehs

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type test struct {
	curr    time.Duration
	attempt int
	next    time.Duration
	done    bool
}

func run(t *testing.T, name string, p RetryPolicy, tests []test) {
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%v_%v", name, test.curr, test.attempt), func(t *testing.T) {
			next, done := p.next(test.curr, test.attempt)
			assert.Equal(t, test.next, next)
			assert.Equal(t, test.done, done)
		})
	}
}

func TestRetryPolicy(t *testing.T) {
	run(t, "unlimitted_attempt", RetryPolicy{Interval: time.Second}, []test{
		{0, 0, 0, false},
		{time.Second, 1, time.Second, false},
		{time.Second, 2, time.Second, false},
		{time.Second, 3, time.Second, false},
	})

	run(t, "single_attempt", RetryPolicy{MaxAttempts: 1}, []test{
		{0, 0, 0, false},
		{0, 1, 0, true},
	})

	run(t, "two_attempts", RetryPolicy{MaxAttempts: 2}, []test{
		{0, 0, 0, false},
		{0, 1, 0, false},
		{0, 2, 0, true},
	})

	run(t, "backoff", RetryPolicy{Interval: time.Second, BackoffCoefficient: 2}, []test{
		{0, 0, 0, false},
		{time.Second, 1, time.Second, false},
		{time.Second, 2, 2 * time.Second, false},
		{2 * time.Second, 3, 4 * time.Second, false},
		{4 * time.Second, 4, 8 * time.Second, false},
	})

	run(t, "backoff_and_cap", RetryPolicy{Interval: time.Second, BackoffCoefficient: 2, MaxInterval: 5 * time.Second}, []test{
		{0, 0, 0, false},
		{time.Second, 1, time.Second, false},
		{time.Second, 2, 2 * time.Second, false},
		{2 * time.Second, 3, 4 * time.Second, false},
		{4 * time.Second, 4, 5 * time.Second, false},
	})
}
