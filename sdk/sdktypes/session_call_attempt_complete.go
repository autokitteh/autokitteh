package sdktypes

import (
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type (
	SessionCallAttemptCompletePB = sessionsv1.Call_Attempt_Complete
	SessionCallAttemptComplete   = *object[*SessionCallAttemptCompletePB]
)

var (
	SessionCallAttemptCompleteFromProto       = makeFromProto(validateSessionCallAttemptComplete)
	StrictSessionCallAttemptCompleteFromProto = makeFromProto(strictValidateSessionCallAttemptComplete)
	ToStrictSessionCallAttemptComplete        = makeWithValidator(strictValidateSessionCallAttemptComplete)
)

func strictValidateSessionCallAttemptComplete(pb *SessionCallAttemptCompletePB) error {
	return validateSessionCallAttemptComplete(pb)
}

func validateSessionCallAttemptComplete(pb *SessionCallAttemptCompletePB) error {
	return validateSessionCallAttemptResult(pb.GetResult())
}

func GetSessionCallAttemptCompleteResult(c SessionCallAttemptComplete) SessionCallAttemptResult {
	if c == nil {
		return nil
	}

	res := c.pb.GetResult()
	if res == nil {
		return nil
	}

	return kittehs.Must1(SessionCallAttemptResultFromProto(res))
}

func NewSessionCallAttemptComplete(last bool, retryInterval time.Duration, result SessionCallAttemptResult) SessionCallAttemptComplete {
	return kittehs.Must1(SessionCallAttemptCompleteFromProto(&SessionCallAttemptCompletePB{
		Result:        result.ToProto(),
		IsLast:        last,
		RetryInterval: durationpb.New(retryInterval),
	}))
}
