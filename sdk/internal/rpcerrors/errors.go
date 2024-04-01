package rpcerrors

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func ToSDKError(err error) error {
	if err == nil {
		return err
	}

	var sdkErr error
	var connectErr *connect.Error

	if !errors.As(err, &connectErr) { // not a connect error?
		return err
	}

	// convert connect errors to sdk. Their strings are almost identical
	switch connectErr.Code() {
	case connect.CodeAlreadyExists:
		sdkErr = sdkerrors.ErrAlreadyExists
	case connect.CodeNotFound:
		sdkErr = sdkerrors.ErrNotFound
	case connect.CodeInvalidArgument:
		sdkErr = sdkerrors.ErrInvalidArgument{Underlying: err}
	case connect.CodeUnimplemented:
		sdkErr = sdkerrors.ErrNotImplemented
	case connect.CodeUnauthenticated:
		sdkErr = sdkerrors.ErrUnauthenticated
	case connect.CodePermissionDenied:
		sdkErr = sdkerrors.ErrUnauthorized
	case connect.CodeResourceExhausted:
		sdkErr = sdkerrors.ErrLimitExceeded
	case connect.CodeUnknown: // returned as connect.Error, but unrelated to RPC, just unwrap underlying error
		return connectErr.Unwrap()
	default:
		sdkErr = sdkerrors.ErrRPC
	}

	// err is a connect error (checked in connect.CodeOf), so we can safely cast it
	connErr := err.(*connect.Error)
	if len(connErr.Details()) != 0 {
		return fmt.Errorf("%w: (%v)", sdkErr, connErr.Details())
	}
	return sdkErr

	/*
	   case connect.CodeAlreadyExists:
	   	err1 = sdkerrors.ErrAlreadyExists
	   // ErrAlreadyExists      = errors.New("already exists")
	   	// 		return

	   case connect.CodeNotFound:
	   	err1 = sdkerrors.ErrNotFound
	   // ErrNotFound           = errors.New("not found")
	   // 		return "not_found"

	   case connect.CodeInvalidArgument:
	   	err1 = sdkerrors.ErrInvalidArgument{Underlying: err}
	   b.WriteString("invalid argument")
	   // 		return "invalid_argument"

	   case connect.CodeUnimplemented:
	   	err1 = sdkerrors.ErrNotImplemented
	   	ErrNotImplemented     = errors.New("not implemented")
	   // 		return "unimplemented"

	   case connect.CodeUnauthenticated:
	   	err1 = sdkerrors.ErrUnauthenticated
	   // ErrUnauthenticated    = errors.New("unauthenticated")
	   // 		return "unauthenticated"

	   case connect.CodePermissionDenied:
	   	err1 = sdkerrors.ErrUnauthorized
	   // ErrUnauthorized       = errors.New("unauthorized")
	   // 		return "permission_denied"

	   case connect.CodeResourceExhausted:
	   	err1 = sdkerrors.ErrLimitExceeded
	   ErrLimitExceeded      = errors.New("limit exceeded")
	   // 		return "resource_exhausted"

	*/
}
