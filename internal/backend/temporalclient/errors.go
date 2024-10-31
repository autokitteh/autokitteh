package temporalclient

import (
	"fmt"

	"go.temporal.io/sdk/temporal"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TranslateError(err error, f string, args ...any) error {
	if err == nil {
		return nil
	}

	return temporal.NewApplicationErrorWithOptions(
		fmt.Sprintf(f, args...),
		sdkerrors.ErrorType(err),
		temporal.ApplicationErrorOptions{
			NonRetryable: !sdkerrors.IsRetryableError(err),
			Cause:        err,
		},
	)
}
