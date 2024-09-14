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

// ToExitCodeError wraps the given error with an OS exit code.
// If the error is nil, it also returns nil.
func ToExitCodeError(err error, whats ...string) error {
	if err == nil {
		return nil
	}

	var code int = GenericFailure

	switch {
	case errors.Is(err, sdkerrors.ErrNotFound):
		// Replace "not found" with "<whats> not found".
		err = fmt.Errorf("%s not found", strings.Join(whats, " "))
		code = NotFoundExitCode
	case errors.As(err, resolver.NotFoundErrorType):
		// Replace "<type> [name] not found" with "<whats> not found".
		err = fmt.Errorf("%s not found", strings.Join(whats, " "))
		code = NotFoundExitCode
	case errors.Is(err, sdkerrors.ErrFailedPrecondition):
		// Replace "failed precondition" with "failed precondition: <whats>".
		err = kittehs.ErrorWithValue(strings.Join(whats, " "), err)
		code = FailedPrecondition
	}

	return NewExitCodeError(code, err)
}

// keep given error, if passed or return notFound if !found condition
func AddNotFoundErrIfCond(err error, found bool) error {
	if err == nil && !found {
		err = sdkerrors.ErrNotFound
	}
	return err
}

// ToExitCodeWithSkipNotFoundFlag returns the given command's error (may be nil) with an OS
// exit code, but considers the "--fail" flag: if set to false, we skip "not found" errors.
func ToExitCodeWithSkipNotFoundFlag(cmd *cobra.Command, err error, whats ...string) error {
	if err == nil {
		return nil
	}

	exitErr := ToExitCodeError(err, whats...).(ExitCodeError) // This cast is always safe.
	if exitErr.Code == NotFoundExitCode {
		flags := cmd.Flags()
		if flags.Lookup("fail") != nil && !kittehs.Must1(flags.GetBool("fail")) {
			return nil
		}
	}

	return exitErr
}
