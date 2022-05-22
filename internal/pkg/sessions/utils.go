package sessions

import (
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	L "github.com/autokitteh/autokitteh/pkg/l"
)

var ErrNoCallValue = errors.New("no call value allowed")

func temporalErrorLogger(l L.L, err error) func(string, ...interface{}) {
	l = l.With("err", err)

	var ae *temporal.ApplicationError
	if errors.As(err, &ae) {
		return l.With("app_error", err).Debug
	}

	return l.Error
}

func withLocalActivityWithoutRetries(ctx workflow.Context) workflow.Context {
	lao := workflow.GetLocalActivityOptions(ctx)
	lao.RetryPolicy = &temporal.RetryPolicy{MaximumAttempts: 1}
	return workflow.WithLocalActivityOptions(ctx, lao)
}

func EnsureNoCallValues(v *apivalues.Value) error {
	return apivalues.Walk(v, func(vv, _ *apivalues.Value, _ apivalues.Role) error {
		if _, ok := vv.Get().(apivalues.CallValue); ok {
			return ErrNoCallValue
		}

		return nil
	})
}

func EnsureNoCallValuesInArgs(vs []*apivalues.Value) error {
	for i, v := range vs {
		if err := EnsureNoCallValues(v); err != nil {
			return fmt.Errorf("%d: %w", i, err)
		}
	}

	return nil
}

func EnsureNoCallValuesInKWArgs(vs map[string]*apivalues.Value) error {
	for k, v := range vs {
		if err := EnsureNoCallValues(v); err != nil {
			return fmt.Errorf("%q: %w", k, err)
		}
	}

	return nil
}
