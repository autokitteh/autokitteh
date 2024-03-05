package values

import (
	"context"
	"fmt"

	"go.starlark.net/starlark"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (vctx *Context) functionToStarlark(v sdktypes.Value) (starlark.Value, error) {
	fv := v.GetFunction()

	fid := fv.UniqueID()

	if fv.ExecutorID().ToRunID() == vctx.RunID {
		// internal function.

		f := vctx.internalFuncs[string(fv.Data())]
		if f == nil {
			return nil, fmt.Errorf("unregistered function id %q", fid)
		}

		return f, nil
	}

	// external function.

	if vctx.externalFuncs == nil {
		vctx.externalFuncs = make(map[string]sdktypes.Value)
	}

	vctx.externalFuncs[fid] = v

	return starlark.NewBuiltin(
		fid,
		func(th *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			akargs, err := kittehs.TransformError(args, vctx.FromStarlarkValue)
			if err != nil {
				return nil, fmt.Errorf("convert values from starlark: %w", err)
			}

			akkwargs := make(map[string]sdktypes.Value, len(kwargs))
			for _, kwarg := range kwargs {
				if len(kwarg) != 2 {
					return nil, fmt.Errorf("invalid kwarg")
				}

				kstr, ok := starlark.AsString(kwarg[0])
				if !ok {
					return nil, fmt.Errorf("expected string key in kwargs, got %v", kwarg[0])
				}

				k, v := kstr, kwarg[1]

				if akkwargs[k], err = vctx.FromStarlarkValue(v); err != nil {
					return nil, fmt.Errorf("convert from starlark kwarg %q: %w", k, err)
				}
			}

			akret, err := vctx.Call(
				context.Background(), // TODO: extract context from thread. Put context in its tls?
				vctx.RunID,
				v,
				akargs,
				akkwargs,
			)
			if err != nil {
				return nil, err
			}

			ret, err := vctx.ToStarlarkValue(akret)
			if err != nil {
				return nil, fmt.Errorf("convert from return value to starlark: %w", err)
			}

			return ret, nil
		},
	), nil
}

func (vctx *Context) fromStarlarkFunction(v *starlark.Function) (sdktypes.Value, error) {
	var sig string

	// check if we already got this function.
	// (v.Hash() wont do because it only hashes its name)
	for k, vv := range vctx.internalFuncs {
		if v == vv {
			sig = k
		}
	}

	if sig == "" {
		vctx.funcSeq++
		sig = fmt.Sprintf("%s#%x", v.Name(), vctx.funcSeq)
	}

	if vctx.internalFuncs == nil {
		vctx.internalFuncs = make(map[string]*starlark.Function)
	}

	vctx.internalFuncs[sig] = v

	argNames := make([]string, v.NumParams())
	for i := 0; i < v.NumParams(); i++ {
		argNames[i], _ = v.Param(i)
	}

	desc, err := sdktypes.ModuleFunctionFromProto(&sdktypes.ModuleFunctionPB{
		Input: kittehs.Transform(argNames, func(s string) *sdktypes.ModuleFunctionFieldPB {
			return &sdktypes.ModuleFunctionFieldPB{Name: s}
		}),
	})
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("invalid function: %w", err)
	}

	return sdktypes.NewFunctionValue(
		sdktypes.NewExecutorID(vctx.RunID),
		v.Name(),
		[]byte(sig),
		nil,
		desc,
	)
}

func (vctx *Context) fromStarlarkBuiltin(b *starlark.Builtin) (sdktypes.Value, error) {
	v, ok := vctx.externalFuncs[b.Name()]
	if !ok {
		return sdktypes.InvalidValue, fmt.Errorf("unregistered external function %q: %w", b.Name(), sdkerrors.ErrNotFound)
	}

	return v, nil
}
