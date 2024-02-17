package sdkmodule

import (
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type varOpts struct {
	desc sdktypes.ModuleVariablePB
	fn   func(xid sdktypes.ExecutorID, data []byte) (sdktypes.Value, error)
}

type VarOpt func(*varOpts) error

func WithVarDoc(doc string) VarOpt {
	return func(opts *varOpts) error {
		opts.desc.Description = doc
		return nil
	}
}

func WithValue(v sdktypes.Value) VarOpt {
	return func(opts *varOpts) error {
		opts.fn = func(sdktypes.ExecutorID, []byte) (sdktypes.Value, error) { return v, nil }
		return nil
	}
}

func WithNewValue(f func(xid sdktypes.ExecutorID, data []byte) (sdktypes.Value, error)) VarOpt {
	return func(opts *varOpts) error {
		opts.fn = f
		return nil
	}
}
