package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type sessionStateTypeTraits struct{}

var _ enumTraits = sessionStateTypeTraits{}

func (sessionStateTypeTraits) Prefix() string           { return "SESSION_STATE_TYPE_" }
func (sessionStateTypeTraits) Names() map[int32]string  { return sessionsv1.SessionStateType_name }
func (sessionStateTypeTraits) Values() map[string]int32 { return sessionsv1.SessionStateType_value }

type SessionStateType struct {
	enum[sessionStateTypeTraits, sessionsv1.SessionStateType]
}

func sessionStateTypeFromProto(e sessionsv1.SessionStateType) SessionStateType {
	return kittehs.Must1(SessionStateTypeFromProto(e))
}

var (
	PossibleSessionStateTypesNames = AllEnumNames[sessionStateTypeTraits]()

	SessionStateTypeUnspecified = sessionStateTypeFromProto(sessionsv1.SessionStateType_SESSION_STATE_TYPE_UNSPECIFIED)
	SessionStateTypeCreated     = sessionStateTypeFromProto(sessionsv1.SessionStateType_SESSION_STATE_TYPE_CREATED)
	SessionStateTypeRunning     = sessionStateTypeFromProto(sessionsv1.SessionStateType_SESSION_STATE_TYPE_RUNNING)
	SessionStateTypeError       = sessionStateTypeFromProto(sessionsv1.SessionStateType_SESSION_STATE_TYPE_ERROR)
	SessionStateTypeCompleted   = sessionStateTypeFromProto(sessionsv1.SessionStateType_SESSION_STATE_TYPE_COMPLETED)
)

func SessionStateTypeFromProto(e sessionsv1.SessionStateType) (SessionStateType, error) {
	return EnumFromProto[SessionStateType](e)
}

func ParseSessionStateType(raw string) (SessionStateType, error) {
	return ParseEnum[SessionStateType](raw)
}
func (e SessionStateType) IsFinal() bool {
	return e.v == sessionsv1.SessionStateType_SESSION_STATE_TYPE_ERROR ||
		e.v == sessionsv1.SessionStateType_SESSION_STATE_TYPE_COMPLETED
}
