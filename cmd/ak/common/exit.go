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
	BadRequest                 = 40
	UnauthenticatedExitCode    = 41
	FailedPreconditionExitCode = 42
	UnauthorizedExitCode       = 43
	NotFoundExitCode           = 44
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

// ToExitCodeError wraps the given error with an OS exit code.
// If the error is nil, it also returns nil.
func ToExitCode(err error) (code int) {
	if err == nil {
		return
	}

	code = GenericFailureExitCode

	var ecerr ExitCodeError

	switch {
	case errors.As(err, &ecerr):
		code = ecerr.Code
	case errors.Is(err, sdkerrors.ErrUnauthorized):
		code = UnauthorizedExitCode
	case errors.Is(err, sdkerrors.ErrUnauthenticated):
		code = UnauthenticatedExitCode
	case errors.Is(err, sdkerrors.ErrNotFound):
		code = NotFoundExitCode
	case errors.As(err, resolver.NotFoundErrorType):
		code = NotFoundExitCode
	case errors.Is(err, sdkerrors.ErrFailedPrecondition):
		code = FailedPreconditionExitCode
	case sdkerrors.IsInvalidArgumentError(err):
		code = BadRequest
	}

	return
}

// WrapErrors wraps the given error with the given strings.
// If the error is nil, it also returns nil.
func WrapError(err error, whats ...string) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, sdkerrors.ErrNotFound):
		// Replace "not found" with "<whats> not found".
		return fmt.Errorf("%s: %w", strings.Join(whats, " "), sdkerrors.ErrNotFound)
	case errors.As(err, resolver.NotFoundErrorType):
		// Replace "<type> [name] not found" with "<whats> not found".
		return fmt.Errorf("%s: %w", strings.Join(whats, " "), sdkerrors.ErrNotFound)
	case errors.Is(err, sdkerrors.ErrFailedPrecondition):
		// Replace "failed precondition" with "failed precondition: <whats>".
		return fmt.Errorf("%w: %s", err, strings.Join(whats, " "))
	default:
		return err
	}
}
