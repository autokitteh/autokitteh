package langtools

import (
	"context"
	"fmt"

	"github.com/autokitteh/autokitteh/sdk/api/apilang"
	"github.com/autokitteh/autokitteh/sdk/api/apiprogram"
	"github.com/autokitteh/autokitteh/sdk/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
	"github.com/autokitteh/autokitteh/pkg/idgen"
)

func RunModule(
	ctx context.Context,
	cat lang.Catalog,
	env *lang.RunEnv,
	mod *apiprogram.Module, // mod must have compiled_code populated.
) (lang.Lang, map[string]*apivalues.Value, *apilang.RunSummary, error) {
	l, err := cat.Acquire(mod.Lang(), env.Scope)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lang %q: %w", mod.Lang(), err)
	}

	gs, sum, err := l.RunModule(ctx, env, mod)

	return l, gs, sum, err
}

func RunModules(
	ctx context.Context,
	cat lang.Catalog,
	env *lang.RunEnv,
	mods []*apiprogram.Module, // first is main
) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
	if len(mods) == 0 {
		return nil, nil, fmt.Errorf("no modules")
	}

	run := func(ctx context.Context, mod *apiprogram.Module) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
		_, vs, sum, err := RunModule(ctx, cat, env, mod)
		return vs, sum, err
	}

	load := env.Load
	env = env.WithStubs()

	if env.Scope == "" {
		// Do not leave this empty. Need to specify scope to allow access to functions
		// returned from one module to another.
		env.Scope = idgen.New("S")
	}

	env.Load = NewInMemoryLoader(mods, load, run)

	return run(ctx, mods[0])
}
