package sdktypes

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
)

type (
	SessionCallSpecPB = sessionsv1.Call_Spec
	SessionCallSpec   = *object[*SessionCallSpecPB]
)

var (
	SessionCallSpecFromProto       = makeFromProto(validateSessionCallSpec)
	StrictSessionCallSpecFromProto = makeFromProto(strictValidateSessionCallSpec)
	ToStrictSessionCallSpec        = makeWithValidator(strictValidateSessionCallSpec)
)

func strictValidateSessionCallSpec(pb *sessionsv1.Call_Spec) error {
	if pb.Function == nil {
		return fmt.Errorf("missing function")
	}

	return validateSessionCallSpec(pb)
}

func validateSessionCallSpec(pb *sessionsv1.Call_Spec) error {
	if err := ValidateValuePB(pb.Function); err != nil {
		return fmt.Errorf("function: %w", err)
	}

	if i, err := kittehs.ValidateList(pb.Args, StrictValidateValuePB); err != nil {
		return fmt.Errorf("arg #%d: %w", i, err)
	}

	if err := kittehs.ValidateMap(pb.Kwargs, func(k string, v *ValuePB) error {
		if _, err := ParseSymbol(k); err != nil {
			return fmt.Errorf("symbol: %w", err)
		}

		if err := StrictValidateValuePB(v); err != nil {
			return fmt.Errorf("value: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("kwargs: %w", err)
	}

	return nil
}

func NewSessionCallSpec(v Value, args []Value, kwargs map[string]Value, seq uint32) SessionCallSpec {
	return kittehs.Must1(StrictSessionCallSpecFromProto(
		&SessionCallSpecPB{
			Function: v.ToProto(),
			Args:     kittehs.Transform(args, ToProto),
			Kwargs:   kittehs.TransformMapValues(kwargs, ToProto),
			Seq:      seq,
		},
	))
}

func GetSessionCallSpecData(v SessionCallSpec) (Value, []Value, map[string]Value) {
	return kittehs.Must1(StrictValueFromProto(v.pb.Function)),
		kittehs.Must1(kittehs.TransformError(v.pb.Args, StrictValueFromProto)),
		kittehs.Must1(kittehs.TransformMapValuesError(v.pb.Kwargs, StrictValueFromProto))
}

func GetSessionCallSpecSeq(v SessionCallSpec) uint32 { return v.pb.Seq }
