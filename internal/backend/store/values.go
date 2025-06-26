package store

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var errTooManyOperands = sdkerrors.NewInvalidArgumentError("too many operands")

type (
	opFn func(curr sdktypes.Value, operands []sdktypes.Value) (set sdktypes.Value, ret sdktypes.Value, _ error)
	op   struct {
		fn    opFn
		read  bool // needs current value from db.
		write bool // should write next value to db.
	}
)

var ops = map[string]op{
	"get": {
		fn: func(v sdktypes.Value, vs []sdktypes.Value) (sdktypes.Value, sdktypes.Value, error) {
			if len(vs) > 0 {
				return sdktypes.InvalidValue, sdktypes.InvalidValue, errTooManyOperands
			}
			return sdktypes.InvalidValue, v, nil
		},
		read: true,
	},
	"set": {
		fn: func(_ sdktypes.Value, vs []sdktypes.Value) (sdktypes.Value, sdktypes.Value, error) {
			if len(vs) == 0 {
				return sdktypes.InvalidValue, sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("missing value to set")
			} else if len(vs) > 1 {
				return sdktypes.InvalidValue, sdktypes.InvalidValue, errTooManyOperands
			}
			return vs[0], sdktypes.Nothing, nil
		},
		write: true,
	},
	"add": {
		fn: func(curr sdktypes.Value, vs []sdktypes.Value) (sdktypes.Value, sdktypes.Value, error) {
			if len(vs) == 0 {
				return sdktypes.InvalidValue, sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("missing value to add")
			} else if len(vs) > 1 {
				return sdktypes.InvalidValue, sdktypes.InvalidValue, errTooManyOperands
			}

			if !curr.IsValid() || curr.IsNothing() {
				// If current value is invalid or nothing, we set it to the given value.
				return vs[0], vs[0], nil
			}

			next, err := sdktypes.AddValues(curr, vs[0])
			if err != nil {
				return sdktypes.InvalidValue, sdktypes.InvalidValue, err
			}

			return next, next, nil
		},
		read:  true,
		write: true,
	},
	"del": {
		// no fn -> next is invalid -> delete on write.
		write: true,
	},
}
