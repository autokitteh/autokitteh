package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	mappingsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/mappings/v1"
)

type MappingPB = mappingsv1.Mapping

type Mapping = *object[*MappingPB]

var (
	MappingFromProto       = makeFromProto(validateMapping)
	StrictMappingFromProto = makeFromProto(strictValidateMapping)
	ToStrictMapping        = makeWithValidator(strictValidateMapping)
)

func strictValidateMapping(pb *mappingsv1.Mapping) error {
	if err := ensureNotEmpty(pb.EnvId, pb.ConnectionId, pb.ModuleName); err != nil {
		return err
	}

	return validateMapping(pb)
}

func validateMapping(pb *mappingsv1.Mapping) error {
	if _, err := ParseMappingID(pb.MappingId); err != nil {
		return fmt.Errorf("mapping id: %w", err)
	}

	if _, err := ParseEnvID(pb.EnvId); err != nil {
		return fmt.Errorf("env id: %w", err)
	}

	if _, err := ParseConnectionID(pb.ConnectionId); err != nil {
		return fmt.Errorf("connection id: %w", err)
	}

	if _, err := kittehs.TransformError(pb.Events, MappingEventFromProto); err != nil {
		return fmt.Errorf("events: %w", err)
	}

	if _, err := ParseSymbol(pb.ModuleName); err != nil {
		return fmt.Errorf("module name: %w", err)
	}

	return nil
}

func GetMappingID(m Mapping) MappingID {
	if m == nil {
		return nil
	}
	return kittehs.Must1(ParseMappingID(m.pb.MappingId))
}

func GetMappingEnvID(m Mapping) EnvID {
	if m == nil {
		return nil
	}
	return kittehs.Must1(ParseEnvID(m.pb.EnvId))
}

func GetMappingConnectionID(m Mapping) ConnectionID {
	if m == nil {
		return nil
	}
	return kittehs.Must1(ParseConnectionID(m.pb.ConnectionId))
}

func GetMappingModuleName(m Mapping) Symbol {
	if m == nil {
		return nil
	}
	return kittehs.Must1(ParseSymbol(m.pb.ModuleName))
}

func GetMappingEvents(m Mapping) []MappingEvent {
	if m == nil {
		return nil
	}
	return kittehs.Must1(kittehs.TransformError(m.pb.Events, MappingEventFromProto))
}
