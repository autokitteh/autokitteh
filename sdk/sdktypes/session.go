package sdktypes

import (
	"errors"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type Session struct {
	object[*SessionPB, SessionTraits]
}

var InvalidSession Session

type SessionPB = sessionv1.Session

type SessionTraits struct{}

func (SessionTraits) Validate(m *SessionPB) error {
	return errors.Join(
		enumField[SessionStateType]("state", m.State),
		idField[DeploymentID]("deployment_id", m.DeploymentId),
		idField[EventID]("event_id", m.EventId),
		idField[SessionID]("parent_session_id", m.ParentSessionId),
		idField[SessionID]("session_id", m.SessionId),
		objectField[CodeLocation]("entrypoint", m.Entrypoint),
		valuesMapField("inputs", m.Inputs),
	)
}

func (SessionTraits) StrictValidate(m *SessionPB) error {
	return errors.Join(
		mandatory("created_at", m.CreatedAt),
		mandatory("deployment_id", m.DeploymentId),
		mandatory("entrypoint", m.Entrypoint),
		mandatory("session_id", m.SessionId),
		mandatory("state", m.State),
	)
}

func SessionFromProto(m *SessionPB) (Session, error) { return FromProto[Session](m) }
func StrictSessionFromProto(m *SessionPB) (Session, error) {
	return Strict(SessionFromProto(m))
}

func (p Session) WithNewID() Session {
	return Session{p.forceUpdate(func(pb *SessionPB) { pb.SessionId = NewSessionID().String() })}
}

func (p Session) ID() SessionID { return kittehs.Must1(ParseSessionID(p.read().SessionId)) }
func (p Session) DeploymentID() DeploymentID {
	return kittehs.Must1(ParseDeploymentID(p.read().DeploymentId))
}
func (p Session) EventID() EventID         { return kittehs.Must1(ParseEventID(p.read().EventId)) }
func (p Session) EntryPoint() CodeLocation { return forceFromProto[CodeLocation](p.read().Entrypoint) }
func (p Session) Memo() map[string]string  { return p.read().Memo }
func (p Session) Inputs() map[string]Value {
	return kittehs.TransformMapValues(p.read().Inputs, forceFromProto[Value])
}

func (p Session) State() SessionStateType {
	return forceEnumFromProto[SessionStateType](p.read().State)
}

func (p Session) WithInputs(inputs map[string]Value) Session {
	return Session{p.forceUpdate(func(pb *SessionPB) { pb.Inputs = kittehs.TransformMapValues(inputs, ToProto) })}
}

func NewSession(deploymentID DeploymentID, parentSessionID SessionID, eventID EventID, ep CodeLocation, inputs map[string]Value, memo map[string]string) Session {
	return kittehs.Must1(SessionFromProto(
		&SessionPB{
			DeploymentId:    deploymentID.String(),
			EventId:         eventID.String(),
			Entrypoint:      ToProto(ep),
			Inputs:          kittehs.TransformMapValues(inputs, ToProto),
			Memo:            memo,
			ParentSessionId: parentSessionID.String(),
		},
	))
}
