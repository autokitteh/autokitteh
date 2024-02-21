package sdktypes

import (
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	triggersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1"
)

type TriggerPB = triggersv1.Trigger

type Trigger = *object[*TriggerPB]

var (
	TriggerFromProto       = makeFromProto(validateTrigger)
	StrictTriggerFromProto = makeFromProto(strictValidateTrigger)
	ToStrictTrigger        = makeWithValidator(strictValidateTrigger)
)

func strictValidateTrigger(pb *triggersv1.Trigger) error {
	if err := ensureNotEmpty(pb.ConnectionId); err != nil {
		return err
	}

	if pb.CodeLocation == nil {
		return errors.New("missing code location")
	}

	return validateTrigger(pb)
}

func validateTrigger(pb *triggersv1.Trigger) error {
	if _, err := ParseTriggerID(pb.TriggerId); err != nil {
		return fmt.Errorf("trigger id: %w", err)
	}

	if _, err := ParseEnvID(pb.EnvId); err != nil {
		return fmt.Errorf("env id: %w", err)
	}

	if _, err := ParseConnectionID(pb.ConnectionId); err != nil {
		return fmt.Errorf("connection id: %w", err)
	}

	if err := validateCodeLocation(pb.CodeLocation); err != nil {
		return fmt.Errorf("code location: %w", err)
	}

	return nil
}

func GetTriggerID(t Trigger) TriggerID {
	if t == nil {
		return nil
	}
	return kittehs.Must1(ParseTriggerID(t.pb.TriggerId))
}

func GetTriggerEnvID(t Trigger) EnvID {
	if t == nil {
		return nil
	}
	return kittehs.Must1(ParseEnvID(t.pb.EnvId))
}

func GetTriggerConnectionID(t Trigger) ConnectionID {
	if t == nil {
		return nil
	}
	return kittehs.Must1(ParseConnectionID(t.pb.ConnectionId))
}

func GetTriggerCodeLocation(t Trigger) CodeLocation {
	if t == nil {
		return nil
	}
	return kittehs.Must1(CodeLocationFromProto(t.pb.CodeLocation))
}

func GetTriggerEventType(t Trigger) string {
	if t == nil {
		return ""
	}
	return t.pb.EventType
}
