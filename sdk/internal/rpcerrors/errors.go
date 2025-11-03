package rpcerrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

// TODO: ENG-2306: fix connect parse when talkint to envoy
// This is a temporary fix for this error
func parseResourceExhaustedError(err *connect.Error) string {
	type connectError struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	var (
		jsonError connectError
		errMsg    = err.Message()
	)

	if parseErr := json.Unmarshal([]byte(err.Message()), &jsonError); parseErr == nil {
		errMsg = jsonError.Message
	}

	return errMsg
}

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
		if strings.Contains(err.Error(), "resource_exhausted") {
			errMsg, sdkErr = parseResourceExhaustedError(connectErr), sdkerrors.ErrResourceExhausted
		} else {
			sdkErr = fmt.Errorf("unknown connect error: %w", connectErr)
		}
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
