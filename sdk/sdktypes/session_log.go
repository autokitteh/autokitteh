package sdktypes

import (
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type (
	SessionLogPB = sessionsv1.SessionLog
	SessionLog   = *object[*SessionLogPB]
)

var (
	SessionLogFromProto       = makeFromProto(validateSessionLog)
	StrictSessionLogFromProto = makeFromProto(strictValidateSessionLog)
	ToStrictSessionLog        = makeWithValidator(strictValidateSessionLog)
)

func strictValidateSessionLog(pb *sessionsv1.SessionLog) error {
	return validateSessionLog(pb)
}

func validateSessionLog(pb *sessionsv1.SessionLog) error {
	if i, err := kittehs.ValidateList(pb.Records, validateSessionLogRecord); err != nil {
		return fmt.Errorf("%d: %w", i, err)
	}

	return nil
}

func NewSessionLogPrint(print string) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		T:    timestamppb.Now(),
		Data: &sessionsv1.SessionLogRecord_Print{Print: print},
	}))
}

func NewSessionLogState(state SessionState) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		T:    timestamppb.Now(),
		Data: &sessionsv1.SessionLogRecord_State{State: state.ToProto()},
	}))
}

func NewSessionLogCallSpec(spec SessionCallSpec) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		T:    timestamppb.Now(),
		Data: &sessionsv1.SessionLogRecord_CallSpec{CallSpec: spec.ToProto()},
	}))
}

func NewSessionLogCallAttemptStart(spec SessionCallAttemptStart) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		T:    timestamppb.Now(),
		Data: &sessionsv1.SessionLogRecord_CallAttemptStart{CallAttemptStart: spec.ToProto()},
	}))
}

func NewSessionLogCallAttemptComplete(spec SessionCallAttemptComplete) SessionLogRecord {
	return kittehs.Must1(SessionLogRecordFromProto(&SessionLogRecordPB{
		T:    timestamppb.Now(),
		Data: &sessionsv1.SessionLogRecord_CallAttemptComplete{CallAttemptComplete: spec.ToProto()},
	}))
}

func NewSessionLog(rs []SessionLogRecord) SessionLog {
	return kittehs.Must1(SessionLogFromProto(&SessionLogPB{
		Records: kittehs.Transform(rs, ToProto),
	}))
}
