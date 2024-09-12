package sdktypes

import sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"

type SessionLogRecordType = sessionv1.SessionLogRecord_Type

const (
	UnspecifiedSessionLogRecordType         SessionLogRecordType = sessionv1.SessionLogRecord_TYPE_UNSPECIFIED
	CallAttemptStartSessionLogRecordType    SessionLogRecordType = sessionv1.SessionLogRecord_TYPE_CALL_ATTEMPT_START
	CallAttemptCompleteSessionLogRecordType SessionLogRecordType = sessionv1.SessionLogRecord_TYPE_CALL_ATTEMPT_COMPLETE
	CallSpecSessionLogRecordType            SessionLogRecordType = sessionv1.SessionLogRecord_TYPE_CALL_SPEC
	StateSessionLogRecordType               SessionLogRecordType = sessionv1.SessionLogRecord_TYPE_STATE
	PrintSessionLogRecordType               SessionLogRecordType = sessionv1.SessionLogRecord_TYPE_PRINT
	StopRequestSessionLogRecordType         SessionLogRecordType = sessionv1.SessionLogRecord_TYPE_STOP_REQUEST
)
