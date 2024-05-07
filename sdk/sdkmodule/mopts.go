package sdkmodule

import (
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type moduleOpts struct {
	funcs map[string]*funcOpts
	vars  map[string]*varOpts
}

type Optfn func(*moduleOpts) error

func ExportFunction(name string, fn sdkexecutor.Function, fopts ...FuncOpt) Optfn {
	return func(mopts *moduleOpts) error {
		if mopts.funcs[name] != nil || mopts.vars[name] != nil {
			return sdkerrors.ErrConflict
		}

		funcOpts := funcOpts{fn: fn}

		for _, fopt := range fopts {
			if err := fopt(&funcOpts); err != nil {
				return err
			}
		}

		mopts.funcs[name] = &funcOpts

		return nil
	}
}

func ExportValue(name string, vopts ...VarOpt) Optfn {
	return func(mopts *moduleOpts) error {
		if _, err := sdktypes.ParseSymbol(name); err != nil {
			return err
		}

		if mopts.vars[name] != nil || mopts.funcs[name] != nil {
			return sdkerrors.ErrConflict
		}

		var varOpts varOpts
		for _, vopt := range vopts {
			if err := vopt(&varOpts); err != nil {
				return err
			}
		}

		if varOpts.fn == nil {
			sdklogger.Panic("value function or constant must be set")
		}

		mopts.vars[name] = &varOpts

		return nil
	}
}
