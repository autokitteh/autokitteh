package sdktypes

import (
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

type SessionStateType sessionsv1.SessionStateType

const (
	UnspecifiedSessionStateType = SessionStateType(sessionsv1.SessionStateType_SESSION_STATE_TYPE_UNSPECIFIED)
	CreatedSessionStateType     = SessionStateType(sessionsv1.SessionStateType_SESSION_STATE_TYPE_CREATED)
	RunningSessionStateType     = SessionStateType(sessionsv1.SessionStateType_SESSION_STATE_TYPE_RUNNING)
	ErrorSessionStateType       = SessionStateType(sessionsv1.SessionStateType_SESSION_STATE_TYPE_ERROR)
	CompletedSessionStateType   = SessionStateType(sessionsv1.SessionStateType_SESSION_STATE_TYPE_COMPLETED)
)

func SessionStateTypeFromProto(s sessionsv1.SessionStateType) (SessionStateType, error) {
	if _, ok := sessionsv1.SessionStateType_name[int32(s.Number())]; ok {
		return SessionStateType(s), nil
	}
	return UnspecifiedSessionStateType, fmt.Errorf("unknown state %v: %w", s, sdkerrors.ErrInvalidArgument)
}

func (s SessionStateType) String() string {
	return strings.TrimPrefix(sessionsv1.SessionStateType_name[int32(s)], "SESSION_STATE_")
}

func (s SessionStateType) ToProto() sessionsv1.SessionStateType {
	return sessionsv1.SessionStateType(s)
}

func ParseSessionStateType(raw string) SessionStateType {
	if raw == "" {
		return UnspecifiedSessionStateType
	}
	upper := strings.ToUpper(raw)
	if !strings.HasPrefix(upper, "SESSION_STATE_TYPE_") {
		upper = "SESSION_STATE_TYPE_" + upper
	}

	st, ok := sessionsv1.SessionStateType_value[upper]
	if !ok {
		return UnspecifiedSessionStateType
	}

	return SessionStateType(st)
}

var PossibleSessionStateTypes = kittehs.Transform(kittehs.MapValuesSortedByKeys(sessionsv1.SessionStateType_name), func(name string) string {
	return strings.TrimPrefix(name, "SESSION_STATE_TYPE_")
})
