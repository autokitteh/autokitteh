package sdktypes

import (
	"fmt"
	"time"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type (
	SessionLogRecordPB = sessionsv1.SessionLogRecord
	SessionLogRecord   = *object[*SessionLogRecordPB]
)

var (
	SessionLogRecordFromProto       = makeFromProto(validateSessionLogRecord)
	StrictSessionLogRecordFromProto = makeFromProto(strictValidateSessionLogRecord)
	ToStrictSessionLogRecord        = makeWithValidator(strictValidateSessionLogRecord)
)

func strictValidateSessionLogRecord(pb *sessionsv1.SessionLogRecord) error {
	return validateSessionLogRecord(pb)
}

func validateSessionLogRecord(pb *sessionsv1.SessionLogRecord) error {
	if _, err := getSessionLogRecordData(pb); err != nil {
		return err
	}

	return nil
}

func NewPrintSessionLogRecord(print string) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		Data: &sessionsv1.SessionLogRecord_Print{
			Print: print,
		},
	}))
}

func NewCallSpecSessionLogRecord(spec SessionCallSpec) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		Data: &sessionsv1.SessionLogRecord_CallSpec{
			CallSpec: spec.ToProto(),
		},
	}))
}

func NewCallAttemptStartSessionLogRecord(start SessionCallAttemptStart) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		Data: &sessionsv1.SessionLogRecord_CallAttemptStart{
			CallAttemptStart: start.ToProto(),
		},
	}))
}

func NewCallAttemptCompleteSessionLogRecord(complete SessionCallAttemptComplete) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		Data: &sessionsv1.SessionLogRecord_CallAttemptComplete{
			CallAttemptComplete: complete.ToProto(),
		},
	}))
}

func NewStateSessionLogRecord(state SessionState) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		Data: &sessionsv1.SessionLogRecord_State{
			State: state.ToProto(),
		},
	}))
}

func GetSessionLogRecordTimestamp(record SessionLogRecord) time.Time {
	return record.pb.T.AsTime()
}

func getSessionLogRecordData(pb *SessionLogRecordPB) (any, error) {
	switch data := pb.Data.(type) {
	case *sessionsv1.SessionLogRecord_Print:
		return data.Print, nil
	case *sessionsv1.SessionLogRecord_CallSpec:
		return SessionCallSpecFromProto(data.CallSpec)
	case *sessionsv1.SessionLogRecord_CallAttemptStart:
		return SessionCallAttemptStartFromProto(data.CallAttemptStart)
	case *sessionsv1.SessionLogRecord_CallAttemptComplete:
		return SessionCallAttemptCompleteFromProto(data.CallAttemptComplete)
	case *sessionsv1.SessionLogRecord_State:
		return SessionStateFromProto(data.State)
	default:
		return nil, fmt.Errorf("unknown session log record data type: %T", data)
	}
}

func GetSessionLogRecordData(r SessionLogRecord) any {
	if r == nil {
		return nil
	}

	return kittehs.Must1(getSessionLogRecordData(r.pb))
}

func GetSessionLogRecordState(r SessionLogRecord) SessionState {
	return GetSessionLogRecordData(r).(SessionState)
}

func GetSessionLogRecordPrint(r SessionLogRecord) string {
	return GetSessionLogRecordData(r).(string)
}
