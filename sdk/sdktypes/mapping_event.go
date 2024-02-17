package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	mappingsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/mappings/v1"
)

type MappingEventPB = mappingsv1.MappingEvent

type MappingEvent = *object[*MappingEventPB]

var (
	MappingEventFromProto       = makeFromProto(validateMappingEvent)
	StrictMappingEventFromProto = makeFromProto(strictValidateMappingEvent)
	ToStrictMappingEvent        = makeWithValidator(strictValidateMappingEvent)
)

func strictValidateMappingEvent(pb *mappingsv1.MappingEvent) error {
	if err := ensureNotEmpty(pb.EventType); err != nil {
		return err
	}

	return validateMappingEvent(pb)
}

func validateMappingEvent(pb *mappingsv1.MappingEvent) error {
	if _, err := CodeLocationFromProto(pb.CodeLocation); err != nil {
		return fmt.Errorf("entrypoint id: %w", err)
	}

	return nil
}

func GetMappingEventType(me MappingEvent) string {
	if me == nil {
		return ""
	}

	return me.pb.EventType
}

func GetMappingEventCodeLocation(me MappingEvent) CodeLocation {
	if me == nil {
		return nil
	}

	return kittehs.Must1(StrictCodeLocationFromProto(me.pb.CodeLocation))
}
