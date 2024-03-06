package rpcerrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func TranslateError(err error) error {
	var err1 error
	switch connect.CodeOf(err) {
	case connect.CodeAlreadyExists:
		err1 = sdkerrors.ErrAlreadyExists
	case connect.CodeNotFound:
		err1 = sdkerrors.ErrNotFound
	case connect.CodeInvalidArgument:
		err1 = sdkerrors.ErrInvalidArgument{Underlying: err}
	case connect.CodeUnimplemented:
		err1 = sdkerrors.ErrNotImplemented
	case connect.CodeUnauthenticated:
		err1 = sdkerrors.ErrUnauthenticated
	case connect.CodePermissionDenied:
		err1 = sdkerrors.ErrUnauthorized
	case connect.CodeResourceExhausted:
		err1 = sdkerrors.ErrLimitExceeded
	default:
		err1 = sdkerrors.ErrRPC
	}

	if cerr := new(connect.Error); errors.As(err, &cerr) {
		return fmt.Errorf("%w: %s (%v)", err1, cerr.Message(), cerr.Details())
	}

	// TODO: This creates double printing of the error code (at least in the CLI). Find a way to get rid of it.
	// eg. "error: create: already exists: already exists: a user with the same handle already exists ([])"
	//                     ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
	return fmt.Errorf("%w: %s", err1, err.Error())
}
