package sdkerrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

var (
	ErrNotImplemented     = errors.New("not implemented")
	ErrAlreadyExists      = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrUnknown            = NewRetryableErrorf("unknown")
	ErrFailedPrecondition = errors.New("failed precondition")
	ErrLimitExceeded      = NewRetryableErrorf("limit exceeded")
	ErrProgram            = errors.New("program error")
)

func IgnoreNotFoundErr[T any](in T, err error) (T, error) {
	var zero T

	if err != nil && !errors.Is(err, ErrNotFound) {
		return zero, err
	}

	return in, nil
}

type RetryableError struct{ Err error }

func (e *RetryableError) Error() string { return fmt.Sprintf("[retryable] %v", e.Err) }
func (e *RetryableError) Unwrap() error { return e.Err }
func NewRetryableError(err error) error {
	return &RetryableError{Err: err}
}

func NewRetryableErrorf(f string, vs ...any) error { return NewRetryableError(fmt.Errorf(f, vs...)) }

func IsRetryableError(err error) bool {
	var r *RetryableError
	return errors.As(err, &r)
}

type ErrInvalidArgument struct {
	Underlying error
}

func (e ErrInvalidArgument) Error() string {
	if e.Underlying != nil {
		return e.Underlying.Error()
	}
	return "invalid argument"
}

func (e ErrInvalidArgument) Unwrap() error { return e.Underlying }

func IsInvalidArgumentError(err error) bool {
	var invalidArg ErrInvalidArgument
	return errors.As(err, &invalidArg)
}

func NewInvalidArgumentError(f string, vs ...any) error {
	return ErrInvalidArgument{Underlying: fmt.Errorf(f, vs...)}
}

// re-wrap sdk as connect error
func AsConnectError(err error) error {
	// in protovalidate Error() is defined on pointer type and there is no error object
	var validationError *protovalidate.ValidationError
	if errors.As(err, &validationError) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	var invalidArg ErrInvalidArgument

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
	switch {
	case errors.Is(err, ErrNotFound):
		return "not_found"
	case IsInvalidArgumentError(err):
		return "invalid_argument"
	case errors.Is(err, ErrUnauthorized):
		return "permission_denied"
	case errors.Is(err, ErrUnauthenticated):
		return "permission_denied"
	case errors.Is(err, ErrAlreadyExists):
		return "already_exists"
	case errors.Is(err, ErrNotImplemented):
		return "not_implemented"
	case errors.Is(err, ErrFailedPrecondition):
		return "failed_precondition"
	case errors.Is(err, ErrLimitExceeded):
		return "limit_exceeded"
	case errors.Is(err, ErrUnknown):
		return "unknown"
	case errors.Is(err, ErrProgram):
		return "program_error"
	default:
		return "unknown_type"
	}
}
