package sdkerrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

var (
	ErrNotImplemented     = errors.New("not implemented")
	ErrRPC                = errors.New("rpc")
	ErrAlreadyExists      = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrUnknown            = errors.New("unknown")
	ErrFailedPrecondition = errors.New("failed precondition")
	ErrLimitExceeded      = errors.New("limit exceeded")
)

func IgnoreNotFoundErr[T any](in T, err error) (T, error) {
	var zero T

	if err != nil && !errors.Is(err, ErrNotFound) {
		return zero, err
	}

	return in, nil
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

func NewInvalidArgumentError(f string, vs ...any) error {
	return ErrInvalidArgument{Underlying: fmt.Errorf(f, vs...)}
}

// re-wrap sdk as connect error
func AsConnectError(err error) error {
	if errors.Is(err, &connect.Error{}) {
		return err
	}

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
	default:
		return connect.NewError(connect.CodeUnknown, err)
	}
}
