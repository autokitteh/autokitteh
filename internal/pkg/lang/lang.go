package lang

import (
	"context"

	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apivalues"
)

type Lang interface {
	IsCompilerVersionSupported(context.Context, string) (bool, error)

	CompileModule(
		_ context.Context,
		path *apiprogram.Path,
		src []byte,
		predecls []string,
	) (*apiprogram.Module, error)

	GetModuleDependencies(context.Context, *apiprogram.Module) ([]*apiprogram.Path, error)

	// In case of cancellation, will return ErrCanceled{Callstack: ...}.
	RunModule(
		ctx context.Context,
		env *RunEnv,
		mod *apiprogram.Module, // mod must have compiled_code populated.
	) (map[string]*apivalues.Value, *apilang.RunSummary, error)

	// In case of cancellation, will return ErrCanceled{Callstack: ...}.
	CallFunction(
		ctx context.Context,
		env *RunEnv,
		fn *apivalues.Value,
		args []*apivalues.Value,
		kws map[string]*apivalues.Value,
	) (*apivalues.Value, *apilang.RunSummary, error)
}
