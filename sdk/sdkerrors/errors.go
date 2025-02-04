package sdkerrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

type (
	TypedError interface{ ErrorType() string }
	typedError struct{ typ string }
)

func newTypedError(t string) error { return typedError{typ: t} }

func (e typedError) Error() string     { return e.typ }
func (e typedError) ErrorType() string { return e.typ }

var (
	ErrNotImplemented     = newTypedError("not_implemented")
	ErrAlreadyExists      = newTypedError("already_exists")
	ErrNotFound           = newTypedError("not_found")
	ErrConflict           = newTypedError("conflict")
	ErrUnauthorized       = newTypedError("unauthorized")
	ErrUnauthenticated    = newTypedError("unauthenticated")
	ErrUnknown            = NewRetryableErrorf("unknown")
	ErrUnretryableUnknown = newTypedError("unretryable_unknown")
	ErrFailedPrecondition = newTypedError("failed_precondition")
	ErrResourceExhausted  = NewRetryableErrorf("resource_exhausted")
	ErrProgram            = newTypedError("program_error")
)

func IgnoreNotFoundErr[T any](in T, err error) (T, error) {
	var zero T

	if err != nil && !errors.Is(err, ErrNotFound) {
		return zero, err
	}

	return in, nil
}

type RetryableError struct {
	Err error
}

func (e *RetryableError) Error() string     { return fmt.Sprintf("[retryable] %v", e.Err) }
func (e *RetryableError) Unwrap() error     { return e.Err }
func (e *RetryableError) ErrorType() string { return "retryable_" + ErrorType(e.Err) }

func NewRetryableError(err error) error {
	return &RetryableError{Err: err}
}

func NewRetryableErrorf(f string, vs ...any) error { return NewRetryableError(fmt.Errorf(f, vs...)) }

func IsRetryableError(err error) bool {
	var r *RetryableError
	return errors.As(err, &r)
}

type InvalidArgumentError struct {
	Underlying error
}

func (e InvalidArgumentError) Error() string {
	if e.Underlying != nil {
		return e.Underlying.Error()
	}
	return "invalid argument"
}

func (e InvalidArgumentError) ErrorType() string { return "invalid_argument" }

func (e InvalidArgumentError) Unwrap() error { return e.Underlying }

func IsInvalidArgumentError(err error) bool {
	var invalidArg InvalidArgumentError
	return errors.As(err, &invalidArg)
}

func NewInvalidArgumentError(f string, vs ...any) error {
	return InvalidArgumentError{Underlying: fmt.Errorf(f, vs...)}
}

// re-wrap sdk as connect error
func AsConnectError(err error) error {
	// in protovalidate Error() is defined on pointer type and there is no error object
	var validationError *protovalidate.ValidationError
	if errors.As(err, &validationError) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	var invalidArg InvalidArgumentError

	switch {
	case errors.Is(err, ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.As(err, &invalidArg):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, ErrUnauthorized):
		return connect.NewError(connect.CodePermissionDenied, err)
	case errors.Is(err, ErrUnauthenticated):
		return connect.NewError(connect.CodeUnauthenticated, err)
	case errors.Is(err, ErrAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, ErrNotImplemented):
		return connect.NewError(connect.CodeUnimplemented, err)
	case errors.Is(err, ErrFailedPrecondition):
		return connect.NewError(connect.CodeFailedPrecondition, err)
	case errors.Is(err, ErrProgram):
		fallthrough
	default:
		return connect.NewError(connect.CodeUnknown, err)
	}
}

func ErrorType(err error) string {
	terr, ok := err.(TypedError)
	if !ok {
		u, ok := err.(interface{ Unwrap() error })
		if ok {
			return ErrorType(u.Unwrap())
		}

		return "unknown"
	}

	return terr.ErrorType()
}
