package sdktypes

import (
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/araddon/dateparse"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func (v Value) ToDuration() (time.Duration, error) {
	switch v := v.Concrete().(type) {
	case DurationValue:
		return v.Value(), nil
	case IntegerValue:
		return time.Second * time.Duration(v.Value()), nil
	case FloatValue:
		return time.Duration(float64(time.Second) * v.Value()), nil
	case StringValue:
		return time.ParseDuration(v.Value())
	default:
		return 0, fmt.Errorf("value not convertible to duration")
	}
}

func (v Value) ToTime() (time.Time, error) {
	switch v := v.Concrete().(type) {
	case TimeValue:
		return v.Value(), nil
	case StringValue:
		return dateparse.ParseAny(v.Value())
	case IntegerValue:
		return time.Unix(v.Value(), 0), nil
	default:
		return time.Time{}, fmt.Errorf("value not convertible to time")
	}
}

func (v Value) ToString() (string, error) {
	switch v := v.Concrete().(type) {
	case StringValue:
		return v.Value(), nil
	case IntegerValue:
		return fmt.Sprintf("%d", v.Value()), nil
	case FloatValue:
		return fmt.Sprintf("%f", v.Value()), nil
	case BooleanValue:
		return fmt.Sprintf("%t", v.Value()), nil
	case DurationValue:
		return v.Value().String(), nil
	case TimeValue:
		return v.Value().String(), nil
	case BytesValue:
		return base64.StdEncoding.EncodeToString(v.Value()), nil
	case SymbolValue:
		return v.Symbol().String(), nil
	default:
		return "", sdkerrors.NewInvalidArgumentError("value not convertible to string")
	}
}

func (v Value) ToStringValuesMap() (map[string]Value, error) {
	switch v := v.Concrete().(type) {
	case DictValue:
		return v.ToStringValuesMap()
	case StructValue:
		return v.Fields(), nil
	case ModuleValue:
		return v.Members(), nil
	default:
		return nil, errors.New("not convertible to map")
	}
}

func (v Value) Unwrap() (any, error)     { return UnwrapValue(v) }
func (v Value) UnwrapInto(dst any) error { return UnwrapValueInto(dst, v) }
