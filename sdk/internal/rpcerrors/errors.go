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

	// convert connect errors to sdk ones. Their strings are almost identical
	errMsg := ""
	switch connectErr.Code() {
	case connect.CodeAlreadyExists:
		sdkErr = sdkerrors.ErrAlreadyExists
	case connect.CodeNotFound:
		sdkErr = sdkerrors.ErrNotFound
	case connect.CodeInvalidArgument:
		sdkErr = sdkerrors.InvalidArgumentError{Underlying: err}
	case connect.CodeUnimplemented:
		sdkErr = sdkerrors.ErrNotImplemented
	case connect.CodeUnauthenticated:
		sdkErr = sdkerrors.ErrUnauthenticated
	case connect.CodePermissionDenied:
		sdkErr = sdkerrors.ErrUnauthorized
	case connect.CodeResourceExhausted:
		sdkErr = sdkerrors.ErrResourceExhausted
		errMsg = connectErr.Message()
	case connect.CodeFailedPrecondition:
		sdkErr = sdkerrors.ErrFailedPrecondition
		errMsg = connectErr.Error()
	case connect.CodeUnknown: // returned as connect.Error, but unrelated to RPC, just unwrap underlying error
		return connectErr.Unwrap()
	default:
		sdkErr = fmt.Errorf("unknown connect error: %w", connectErr)
	}

	// err is a connect error (checked in connect.CodeOf), so we can safely cast it
	if len(connectErr.Details()) != 0 {
		errMsg += fmt.Sprintf(" (%v)", connectErr.Details())
	}
	if len(errMsg) != 0 {
		return fmt.Errorf("%w: %s", sdkErr, errMsg)
	}
	return sdkErr
}
