package sdktypes

import (
	"fmt"

	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

type (
	SessionPB = sessionsv1.Session
	Session   = *object[*SessionPB]
)

var (
	SessionFromProto       = makeFromProto(validateSession)
	StrictSessionFromProto = makeFromProto(strictValidateSession)
	ToStrictSession        = makeWithValidator(strictValidateSession)
)

func strictValidateSession(pb *sessionsv1.Session) error {
	if err := ensureNotEmpty(pb.SessionId, pb.DeploymentId); err != nil {
		return err
	}

	return validateSession(pb)
}

func validateSession(pb *sessionsv1.Session) error {
	if _, err := ParseSessionID(pb.SessionId); err != nil {
		return err
	}

	if _, err := ParseSessionID(pb.ParentSessionId); err != nil {
		return err
	}

	if _, err := ParseDeploymentID(pb.DeploymentId); err != nil {
		return err
	}

	if _, err := ParseEventID(pb.EventId); err != nil {
		return err
	}

	if _, err := CodeLocationFromProto(pb.Entrypoint); err != nil {
		return err
	}

	if err := kittehs.ValidateMap(pb.Inputs, func(k string, v *ValuePB) error {
		if _, err := ParseSymbol(k); err != nil {
			return err
		}

		if err := ValidateValuePB(v); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to validate session inputs: %w", err)
	}

	return nil
}

func GetSessionID(e Session) SessionID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseSessionID(e.pb.SessionId))
}

func GetParentSessionID(e Session) SessionID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseSessionID(e.pb.ParentSessionId))
}

func GetSessionDeploymentID(e Session) DeploymentID {
	if e == nil {
		return nil
	}
	return kittehs.Must1(ParseDeploymentID(e.pb.DeploymentId))
}

func GetSessionEventID(e Session) EventID {
	if e == nil || e.pb.EventId == "" {
		return nil
	}
	return kittehs.Must1(ParseEventID(e.pb.EventId))
}

func GetSessionEntryPoint(e Session) CodeLocation {
	if e == nil {
		return nil
	}
	return kittehs.Must1(CodeLocationFromProto(e.pb.Entrypoint))
}

func GetSessionInputs(e Session) map[string]Value {
	if e == nil {
		return nil
	}

	return kittehs.Must1(kittehs.TransformMapValuesError(e.pb.Inputs, ValueFromProto))
}

func GetSessionMemo(e Session) map[string]string {
	if e == nil {
		return nil
	}

	return e.pb.Memo
}

func GetSessionLatestState(e Session) SessionStateType {
	if e == nil {
		return UnspecifiedSessionStateType
	}

	return SessionStateType(e.pb.State)
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
