package common

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func AddFailIfNotFoundFlag(cmd *cobra.Command) {
	cmd.Flags().BoolP("fail", "f", false, "fail if not found")
}

func AddFailIfError(cmd *cobra.Command) {
	cmd.Flags().BoolP("fail", "f", false, "fail on error")
}

func FailIfNotFound(cmd *cobra.Command, what string, found bool) error {
	if !found {
		return FailNotFound(cmd, what)
	}
	return nil
}

func FailNotFound(cmd *cobra.Command, what string) error {
	if kittehs.Must1(cmd.Flags().GetBool("fail")) {
		return NewExitCodeError(NotFoundExitCode, fmt.Errorf("%s not found", what))
	}
	return nil
}

func ToExitCodeError(err error, what string) error {
	if err == nil {
		return nil
	}
	msg := what
	var code int = GenericFailure
	switch {
	case errors.Is(err, sdkerrors.ErrNotFound):
		return NewExitCodeError(NotFoundExitCode, fmt.Errorf("%s not found", what))
	case errors.Is(err, sdkerrors.ErrFailedPrecondition):
		msg = fmt.Sprintf("on %s", what)
		code = FailedPrecondition
	case errors.As(err, resolver.NotFoundErrorType):
		return NewExitCodeError(NotFoundExitCode, fmt.Errorf("%s not found", what))
	}
	return NewExitCodeError(code, fmt.Errorf("%w: %s", err, msg))
}

func FailIfError(cmd *cobra.Command, err error, what string) error {
	if kittehs.Must1(cmd.Flags().GetBool("fail")) && err != nil {
		return ToExitCodeError(err, what)
	}
	return nil
}
