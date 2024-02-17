package sdktypes

import (
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type (
	SessionCallPB = sessionsv1.Call
	SessionCall   = *object[*SessionCallPB]
)

var (
	SessionCallFromProto       = makeFromProto(validateSessionCall)
	StrictSessionCallFromProto = makeFromProto(strictValidateSessionCall)
	ToStrictSessionCall        = makeWithValidator(strictValidateSessionCall)
)

func strictValidateSessionCall(pb *sessionsv1.Call) error {
	// TODO
	return validateSessionCall(pb)
}

func validateSessionCall(pb *sessionsv1.Call) error {
	// TODO
	return nil
}
