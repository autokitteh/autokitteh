package sdktypes

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func AddValues(a, b Value) (Value, error) {
	if a.IsInteger() {
		i, err := b.ToInt64()
		if err != nil {
			return InvalidValue, err
		}

		return NewIntegerValue(a.GetInteger().Value() + i), nil
	}

	if a.IsFloat() {
		f, err := b.ToFloat64()
		if err != nil {
			return InvalidValue, err
		}

		return NewFloatValue(a.GetFloat().Value() + f), nil
	}

	if a.IsDuration() {
		d, err := b.ToDuration()
		if err != nil {
			return InvalidValue, err
		}

		return NewDurationValue(a.GetDuration().Value() + d), nil
	}

	return InvalidValue, sdkerrors.NewInvalidArgumentError("cannot add values of type %s and %s", a.Type(), b.Type())
}
