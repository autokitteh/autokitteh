package sdkerrors

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/bufbuild/protovalidate-go"
)

var (
	ErrNotImplemented     = errors.New("not implemented")
	ErrRPC                = errors.New("rpc")
	ErrAlreadyExists      = errors.New("already exists")
	ErrNotFound           = errors.New("not found")
	ErrInvalidArgument    = errors.New("invalid argument")
	ErrConflict           = errors.New("conflict")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrUnauthenticated    = errors.New("unauthenticated")
	ErrUnknown            = errors.New("unknown")
	ErrFailedPrecondition = errors.New("failed precondition")
)

func IgnoreNotFoundErr[T any](t *T, err error) (*T, error) {
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	if t == nil {
		return nil, nil
	}

	return t, nil
}

// re-wrap sdk as connect error
func AsConnectError(err error) error {
	// in protovalidate Error() is defined on pointer type and there is no error object
	var validationError *protovalidate.ValidationError
	if errors.As(err, &validationError) {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, ErrInvalidArgument):
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
