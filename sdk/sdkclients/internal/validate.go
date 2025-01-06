package internal

import (
	"google.golang.org/protobuf/proto"

	akproto "go.autokitteh.dev/autokitteh/proto"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func Validate(pb proto.Message) error {
	if err := akproto.Validate(pb); err != nil {
		return sdkerrors.NewInvalidArgumentError("invalid proto message: %v", err)
	}

	return nil
}
