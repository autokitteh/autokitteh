package sdktypes

import (
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type (
	SessionCallAttemptStartPB = sessionsv1.Call_Attempt_Start
	SessionCallAttemptStart   = *object[*SessionCallAttemptStartPB]
)

var (
	SessionCallAttemptStartFromProto       = makeFromProto(validateSessionCallAttemptStart)
	StrictSessionCallAttemptStartFromProto = makeFromProto(strictValidateSessionCallAttemptStart)
	ToStrictSessionCallAttemptStart        = makeWithValidator(strictValidateSessionCallAttemptStart)
)

func strictValidateSessionCallAttemptStart(pb *SessionCallAttemptStartPB) error {
	return validateSessionCallAttemptStart(pb)
}

func validateSessionCallAttemptStart(pb *SessionCallAttemptStartPB) error {
	return nil
}
