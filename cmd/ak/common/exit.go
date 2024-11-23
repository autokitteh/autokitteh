package common

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

// Must correspond with `tests/systems/captures.go`.
const (
	GenericFailureExitCode     = 1
	NotFoundExitCode           = 44
	NotAMemberExitCode         = NotFoundExitCode
	FailedPreconditionExitCode = 42
	UnauthroizedExitCode       = 43
	UnauthenticatedExitCode    = 41
)

type ExitCodeError struct {
	Err  error
	Code int
}

func (e ExitCodeError) Error() string {
	return e.Err.Error()
}

var _ error = ExitCodeError{}

func NewExitCodeError(code int, err error) ExitCodeError {
	return ExitCodeError{Err: err, Code: code}
}

func Exit(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(ToExitCode(err))
}

func ToExitCode(err error) (code int) {
	if err == nil {
		return
	}

	code = GenericFailureExitCode

	var ecerr ExitCodeError
	if errors.As(err, &ecerr) {
		code = ecerr.Code
	} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
		code = UnauthroizedExitCode
	} else if errors.Is(err, sdkerrors.ErrUnauthenticated) {
		code = UnauthenticatedExitCode
	} else if errors.Is(err, sdkerrors.ErrNotFound) {
		code = NotFoundExitCode
	} else if errors.As(err, resolver.NotFoundErrorType) {
		code = NotFoundExitCode
	} else if errors.Is(err, sdkerrors.ErrFailedPrecondition) {
		code = FailedPreconditionExitCode
	}

	return
}

// ToExitCodeError wraps the given error with an OS exit code.
// If the error is nil, it also returns nil.
func WrapError(err error, whats ...string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, sdkerrors.ErrNotFound):
		// Replace "not found" with "<whats> not found".
		return fmt.Errorf("%s not found", strings.Join(whats, " "))
	case errors.As(err, resolver.NotFoundErrorType):
		// Replace "<type> [name] not found" with "<whats> not found".
		return fmt.Errorf("%s not found", strings.Join(whats, " "))
	case errors.Is(err, sdkerrors.ErrFailedPrecondition):
		// Replace "failed precondition" with "failed precondition: <whats>".
		return fmt.Errorf("%w: %s", err, strings.Join(whats, " "))
	default:
		return err
	}
}
