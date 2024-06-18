package runtime

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/tls"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/values"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type starlarkBuiltinFunc = func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error)

var (
	builtins = kittehs.TransformMap(builinsFuncs, func(k string, v starlarkBuiltinFunc) (string, starlark.Value) {
		return k, starlark.NewBuiltin(k, v)
	})

	builinsFuncs = map[string]starlarkBuiltinFunc{
		"run_activity": runActivityBuiltinFunc,
		"catch":        catchBuiltinFunc,
		"fail":         failBuiltinFunc,
		"globals":      globalsBuiltinFunc,
		"module":       starlarkstruct.MakeModule,
		"struct":       starlarkstruct.Make,
	}
)

func runActivityBuiltinFunc(th *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("missing function argument")
	}

	tlsContext := tls.Get(th)
	if tlsContext == nil {
		return nil, fmt.Errorf("context is not set")
	}

	vctx := values.FromTLS(th)
	if vctx == nil {
		return nil, fmt.Errorf("value context is not set")
	}

	akV, err := vctx.FromStarlarkValue(args[0])
	if err != nil {
		return nil, fmt.Errorf("cannot convert function from starlark: %w", err)
	}

	akArgs, err := kittehs.TransformError(args[1:], vctx.FromStarlarkValue)
	if err != nil {
		return nil, fmt.Errorf("cannot convert args from starlark: %w", err)
	}

	akKwArgs, err := kittehs.ListToMapError(kwargs, func(t starlark.Tuple) (string, sdktypes.Value, error) {
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
		return nil, err
	}

	// force activity call.
	akV = akV.WithoutFunctionFlag(sdktypes.PureFunctionFlag)

	rv, err := tlsContext.Callbacks.Call(tlsContext.GoCtx, tlsContext.RunID, akV, akArgs, akKwArgs)
	if err != nil {
		return nil, err
	}

	return vctx.ToStarlarkValue(rv)
}

func catchBuiltinFunc(th *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// TODO: optionally do not catch starlark script syntax errors, just actually failures.

	fn, ok := args[0].(starlark.Callable)
	if !ok {
		return nil, errors.New("first argument must be a function")
	}

	vctx := values.FromTLS(th)
	if vctx == nil {
		return nil, fmt.Errorf("value context is not set")
	}

	value, err := starlark.Call(th, fn, args[1:], kwargs)
	if err != nil {
		if perr, ok := sdktypes.FromError(err); ok {
			if slv, cerr := vctx.ToStarlarkValue(perr.Value()); cerr == nil {
				return starlark.Tuple{starlark.None, slv}, nil
			}
		}

		return starlark.Tuple{starlark.None, starlark.String(err.Error())}, nil
	}

	return starlark.Tuple{value, starlark.None}, nil
}

func failBuiltinFunc(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	vctx := values.FromTLS(th)
	if vctx == nil {
		return nil, fmt.Errorf("value context is not set")
	}

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

		v, err := sdktypes.NewListValue(vs)
		if err != nil {
			return nil, fmt.Errorf("cannot create list value: %w", err)
		}
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
}

func globalsBuiltinFunc(th *starlark.Thread, bi *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(bi.Name(), args, kwargs); err != nil {
		return nil, err
	}

	tlsContext := tls.Get(th)
	if tlsContext == nil {
		return nil, fmt.Errorf("context is not set")
	}

	d := starlark.NewDict(len(tlsContext.Globals))
	for k, v := range tlsContext.Globals {
		kittehs.Must0(d.SetKey(starlark.String(k), v))
	}

	return d, nil
}
