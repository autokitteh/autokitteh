package sdkmodule

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Module interface {
	Describe() sdktypes.Module

	Configure(ctx context.Context, xid sdktypes.ExecutorID, cid sdktypes.ConnectionID) (map[string]sdktypes.Value, error)

	sdkexecutor.Caller
}

type module struct {
	desc sdktypes.Module
	opts moduleOpts
}

func New(optfns ...Optfn) Module {
	opts := moduleOpts{
		funcs: make(map[string]*funcOpts),
		vars:  make(map[string]*varOpts),
	}

	for i, optfn := range optfns {
		if err := optfn(&opts); err != nil {
			sdklogger.Panic("option error", "i", i, "err", err)
		}
	}

	return &module{
		desc: kittehs.Must1(sdktypes.NewModule(
			kittehs.TransformMapValues(opts.funcs, func(f *funcOpts) sdktypes.ModuleFunction {
				return kittehs.Must1(sdktypes.ModuleFunctionFromProto(&f.desc))
			}),
			kittehs.TransformMapValues(opts.vars, func(v *varOpts) sdktypes.ModuleVariable {
				return kittehs.Must1(sdktypes.ModuleVariableFromProto(&v.desc))
			}),
		)),
		opts: opts,
	}
}

func (m *module) Configure(ctx context.Context, xid sdktypes.ExecutorID, cid sdktypes.ConnectionID) (map[string]sdktypes.Value, error) {
	values := make(map[string]sdktypes.Value)

	var data []byte
	if cid.IsValid() {
		data = []byte(cid.String())
	}

	for k, v := range m.opts.vars {
		if values[k].IsValid() {
			return nil, fmt.Errorf("value name %q is already set: %w", k, sdkerrors.ErrConflict)
		}

		var err error
		if values[k], err = v.fn(xid, data); err != nil {
			return nil, fmt.Errorf("%w: %s", err, k)
		}
	}

	for k, f := range m.opts.funcs {
		var err error
		values[k], err = sdktypes.NewFunctionValue(xid, k, data, f.flags, kittehs.Must1(sdktypes.ModuleFunctionFromProto(&f.desc)))
		if err != nil {
			return nil, fmt.Errorf("%w: %s", err, k)
		}
	}

	return values, nil
}

func (m *module) Call(ctx context.Context, fnv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	name := fnv.GetFunction().Name()
	if !name.IsValid() {
		return sdktypes.InvalidValue, sdkerrors.ErrInvalidArgument{}
	}

	fn, ok := m.opts.funcs[name.String()]
	if !ok {
		return sdktypes.InvalidValue, sdkerrors.ErrNotFound
	}

	return fn.fn(wrapCallContext(ctx, m, fnv), args, kwargs)
}

func (m *module) Describe() sdktypes.Module { return m.desc }
