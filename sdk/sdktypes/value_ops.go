package sdktypes

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func AddValues(a, b Value) (Value, error) {
	if a.IsBigInteger() || b.IsBigInteger() {
		aa, err := a.ToBigInteger()
		if err != nil {
			return InvalidValue, err
		}

		bb, err := b.ToBigInteger()
		if err != nil {
			return InvalidValue, err
		}

		_ = aa.Add(aa, bb)

		if aa.IsInt64() {
			return NewIntegerValue(aa.Int64()), nil
		}

		if aa.IsUint64() {
			return NewIntegerValue(aa.Uint64()), nil
		}

		return NewBigIntegerValue(aa), nil
	}

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
