package store

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var errTooManyOperands = sdkerrors.NewInvalidArgumentError("too many operands")

func mutateValue(v sdktypes.Value, op string, operands ...sdktypes.Value) (set sdktypes.Value, ret sdktypes.Value, _ error) {
	f, ok := ops[op]
	if !ok {
		return sdktypes.InvalidValue, sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("unknown operation")
	}

	if !v.IsValid() {
		v = sdktypes.Nothing
	}

	return f(v, operands)
}

var ops = map[string]func(sdktypes.Value, []sdktypes.Value) (set sdktypes.Value, ret sdktypes.Value, _ error){
	"get": func(v sdktypes.Value, vs []sdktypes.Value) (sdktypes.Value, sdktypes.Value, error) {
		if len(vs) > 0 {
			return sdktypes.InvalidValue, sdktypes.InvalidValue, errTooManyOperands
		}
		return v, v, nil
	},
	"set": func(_ sdktypes.Value, vs []sdktypes.Value) (sdktypes.Value, sdktypes.Value, error) {
		if len(vs) == 0 {
			return sdktypes.InvalidValue, sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("missing value to set")
		} else if len(vs) > 1 {
			return sdktypes.InvalidValue, sdktypes.InvalidValue, errTooManyOperands
		}
		return vs[0], sdktypes.Nothing, nil
	},
	"del": func(_ sdktypes.Value, vs []sdktypes.Value) (sdktypes.Value, sdktypes.Value, error) {
		if len(vs) > 0 {
			return sdktypes.InvalidValue, sdktypes.InvalidValue, errTooManyOperands
		}
		return sdktypes.InvalidValue, sdktypes.Nothing, nil
	},
}
