package common

import (
	"errors"
	"fmt"
	"strings"

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
		return ToExitCodeWithSkipNotFoundFlag(cmd, sdkerrors.ErrNotFound, what)
	}
	return nil
}

func ToExitCodeErrorNotNilErr(err error, whats ...string) ExitCodeError {
	msg := strings.Join(whats, " ")
	var code int = GenericFailure

	switch {
	case errors.Is(err, sdkerrors.ErrNotFound):
		return NewExitCodeError(NotFoundExitCode, fmt.Errorf("%s not found", msg))
	case errors.Is(err, sdkerrors.ErrFailedPrecondition):
		msg = fmt.Sprintf("on %s", msg)
		code = FailedPrecondition
	case errors.As(err, resolver.NotFoundErrorType):
		return NewExitCodeError(NotFoundExitCode, fmt.Errorf("%s not found", msg))
	}
	if msg == "" {
		return NewExitCodeError(code, err)
	}
	// return NewExitCodeError(code, fmt.Errorf("%w: %s", err, msg))
	return NewExitCodeError(code, err)
}

func ToExitCodeError(err error, whats ...string) error {
	if err == nil {
		return nil
	}
	return ToExitCodeErrorNotNilErr(err, whats...)
}

// keep given error, if passed or return notFound if !found condition
func AddNotFoundErrIfCond(err error, found bool) error {
	if err == nil && !found {
		err = sdkerrors.ErrNotFound
	}
	return err
}

// FIXME: introduced to avoid more changes in existing code due non-skipping of NotFoundExitCode in services
// Need to be combined and refactored with ToExitCodeWithSkipNotFoundFlag and FailIfNotFound
func ToExitCodeWithSkipNotFoundFlag(cmd *cobra.Command, err error, whats ...string) error {
	if err == nil {
		return nil
	}
	exitErr := ToExitCodeErrorNotNilErr(err, whats...)
	if exitErr.Code == NotFoundExitCode {
		flags := cmd.Flags()
		if flags.Lookup("fail") != nil && !kittehs.Must1(flags.GetBool("fail")) {
			return nil
		}
	}
	return exitErr
}
