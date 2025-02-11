package sdktypes

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type SessionLogRecord struct {
	object[*SessionLogRecordPB, SessionLogRecordTraits]
}

func init() { registerObject[SessionLogRecord]() }

var InvalidSessionLogRecord SessionLogRecord

type SessionLogRecordPB = sessionv1.SessionLogRecord

type SessionLogRecordTraits struct{ immutableObjectTrait }

func (SessionLogRecordTraits) Validate(m *SessionLogRecordPB) error {
	return errors.Join(
		objectField[SessionCallAttemptStart]("call_attempt_start", m.CallAttemptStart),
		objectField[SessionCallAttemptComplete]("call_attempt_complete", m.CallAttemptComplete),
		objectField[SessionCallSpec]("call_spec", m.CallSpec),
		objectField[SessionState]("state", m.State),
	)
}

func (SessionLogRecordTraits) StrictValidate(m *SessionLogRecordPB) error {
	return errors.Join(
		mandatory("t", m.T),
		oneOfMessage(m /* ignore: */, "t", "process_id"),
	)
}

func SessionLogRecordFromProto(m *SessionLogRecordPB) (SessionLogRecord, error) {
	return FromProto[SessionLogRecord](m)
}

func StrictSessionLogRecordFromProto(m *SessionLogRecordPB) (SessionLogRecord, error) {
	return Strict(SessionLogRecordFromProto(m))
}

func (s SessionLogRecord) GetPrint() (Value, bool) {
	if m := s.read(); m.Print != nil {
		if m.Print.Value == nil {
			return NewStringValue(m.Print.Text), true
		}

		return kittehs.Must1(ValueFromProto(m.Print.Value)), true
	}

	return InvalidValue, false
}

func (s SessionLogRecord) GetCallSpec() SessionCallSpec {
	return forceFromProto[SessionCallSpec](s.read().CallSpec)
}

func (s SessionLogRecord) GetState() SessionState {
	return forceFromProto[SessionState](s.read().State)
}

func (s SessionLogRecord) GetStopRequest() (string, bool) {
	if m := s.read(); m.StopRequest != nil {
		return m.StopRequest.Reason, true
	}

	return "", false
}

func NewPrintSessionLogRecord(t time.Time, v Value, callSeq uint32) SessionLogRecord {
	var text string
	if v.IsString() {
		text = v.GetString().Value()
	} else {
		text = v.String()
	}

	return forceFromProto[SessionLogRecord](&SessionLogRecordPB{
		T: timestamppb.New(t),
		Print: &sessionv1.SessionLogRecord_Print{
			Text:    text,
			Value:   v.ToProto(),
			CallSeq: callSeq,
		},
	})
}

func NewStopRequestSessionLogRecord(t time.Time, reason string) SessionLogRecord {
	return forceFromProto[SessionLogRecord](&SessionLogRecordPB{
		T:           timestamppb.New(t),
		StopRequest: &sessionv1.SessionLogRecord_StopRequest{Reason: reason},
	})
}

func NewStateSessionLogRecord(t time.Time, state SessionState) SessionLogRecord {
	if !state.IsValid() {
		return InvalidSessionLogRecord
	}

	return forceFromProto[SessionLogRecord](&SessionLogRecordPB{
		T:     timestamppb.New(t),
		State: state.ToProto(),
	})
}

func NewCallAttemptStartSessionLogRecord(t time.Time, s SessionCallAttemptStart) SessionLogRecord {
	if !s.IsValid() {
		return InvalidSessionLogRecord
	}

	return forceFromProto[SessionLogRecord](&SessionLogRecordPB{
		T:                timestamppb.New(t),
		CallAttemptStart: s.ToProto(),
	})
}

func NewCallAttemptCompleteSessionLogRecord(t time.Time, s SessionCallAttemptComplete) SessionLogRecord {
	if !s.IsValid() {
		return InvalidSessionLogRecord
	}

	return forceFromProto[SessionLogRecord](&SessionLogRecordPB{
		T:                   timestamppb.New(t),
		CallAttemptComplete: s.ToProto(),
	})
}

func NewCallSpecSessionLogRecord(t time.Time, s SessionCallSpec) SessionLogRecord {
	if !s.IsValid() {
		return InvalidSessionLogRecord
	}

	return forceFromProto[SessionLogRecord](&SessionLogRecordPB{
		T:        timestamppb.New(t),
		CallSpec: s.ToProto(),
	})
}

func (r SessionLogRecord) WithoutTimestamp() SessionLogRecord {
	m := r.read()
	m.T = nil

	if m.CallAttemptStart != nil {
		m.CallAttemptStart.StartedAt = nil
	}

	if m.CallAttemptComplete != nil {
		m.CallAttemptComplete.CompletedAt = nil
	}

	return forceFromProto[SessionLogRecord](m)
}

func (r SessionLogRecord) Timestamp() time.Time {
	return r.read().T.AsTime()
}

func (r SessionLogRecord) WithProcessID(pid string) SessionLogRecord {
	m := r.read()
	m.ProcessId = pid
	return forceFromProto[SessionLogRecord](m)
}
