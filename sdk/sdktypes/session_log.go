package sdktypes

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionLog struct {
	object[*SessionLogPB, SessionLogTraits]
}

func init() { registerObject[SessionLog]() }

var InvalidSessionLog SessionLog

type SessionLogPB = sessionv1.SessionLog

type SessionLogTraits struct{ immutableObjectTrait }

func (SessionLogTraits) Validate(m *SessionLogPB) error {
	return objectsSlice[SessionLogRecord](m.Records)
}
func (SessionLogTraits) StrictValidate(m *SessionLogPB) error { return nil }

func SessionLogFromProto(m *SessionLogPB) (SessionLog, error) { return FromProto[SessionLog](m) }
func StrictSessionLogFromProto(m *SessionLogPB) (SessionLog, error) {
	return Strict(SessionLogFromProto(m))
}

func NewSessionLog(rs []SessionLogRecord) SessionLog {
	return forceFromProto[SessionLog](&SessionLogPB{Records: kittehs.Transform(rs, ToProto)})
}

func (l SessionLog) Records() []SessionLogRecord {
	return kittehs.Transform(l.read().Records, forceFromProto[SessionLogRecord])
}
