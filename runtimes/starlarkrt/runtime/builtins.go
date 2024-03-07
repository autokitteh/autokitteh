package runtime

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/values"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var builtins = map[string]func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error){
	"catch": func(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// TODO: optionally do not catch starlark script syntax errors, just actually failures.

		fn, ok := args[0].(starlark.Callable)
		if !ok {
			return nil, errors.New("first argument must be a function")
		}

		value, err := starlark.Call(thread, fn, args[1:], kwargs)
		if err != nil {
			if perr, ok := sdktypes.FromError(err); ok {
				v := perr.Value()
				vctx := values.Context{}
				if slv, cerr := vctx.ToStarlarkValue(v); cerr == nil {
					return starlark.Tuple{starlark.None, slv}, nil
				}
			}

			return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
		}

		return starlark.Tuple{value, starlark.None}, nil
	},
	"fail": func(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		vctx := values.Context{}

		if len(args) != 0 && len(kwargs) != 0 {
			return nil, fmt.Errorf("cannot specify both positional and keyword arguments")
		}

		if len(args) == 1 {
			v, err := vctx.FromStarlarkValue(args[0])
			if err != nil {
				return nil, fmt.Errorf("cannot convert value from starlark: %w", err)
			}
			return nil, sdktypes.NewProgramError(v, nil, nil).ToError()
		}

		if len(args) > 1 {
			vs, err := kittehs.TransformError(args, vctx.FromStarlarkValue)
			if err != nil {
				return nil, fmt.Errorf("cannot convert values from starlark: %w", err)
			}

			v := sdktypes.NewListValue(vs)
			return nil, sdktypes.NewProgramError(v, nil, nil).ToError()
		}

		if len(kwargs) == 0 {
			return nil, sdktypes.NewProgramError(sdktypes.NewStringValue("user triggered error"), nil, nil).ToError()
		}

		vs, err := kittehs.ListToMapError(kwargs, func(t starlark.Tuple) (string, sdktypes.Value, error) {
			if len(t) != 2 {
				return "", sdktypes.InvalidValue, fmt.Errorf("expected 2 elements, got %d", len(t))
			}

			k, ok := t[0].(starlark.String)
			if !ok {
				return "", sdktypes.InvalidValue, fmt.Errorf("expected string, got %T", t[0])
			}

			v, err := vctx.FromStarlarkValue(t[1])
			if err != nil {
				return "", sdktypes.InvalidValue, fmt.Errorf("cannot convert value from starlark: %w", err)
			}

			return k.GoString(), v, nil
		})
		if err != nil {
			return nil, fmt.Errorf("cannot convert values from starlark: %w", err)
		}

		v := sdktypes.NewDictValueFromStringMap(vs)
		return nil, sdktypes.NewProgramError(v, nil, nil).ToError()
	},
	"globals": func(thread *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		if err := starlark.UnpackArgs(bi.Name(), args, kwargs); err != nil {
			return nil, err
		}

		globals, ok := thread.Local(globalsTLSKey).(starlark.StringDict)
		if !ok {
			return nil, fmt.Errorf("globals are not set")
		}

		d := starlark.NewDict(len(globals))
		for k, v := range globals {
			kittehs.Must0(d.SetKey(starlark.String(k), v))
		}

		return d, nil
	},
}
