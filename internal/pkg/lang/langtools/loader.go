package langtools

import (
	"context"
	"errors"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apilang"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiprogram"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/lang"
)

func NewInMemoryLoader(
	mods []*apiprogram.Module,
	load lang.LoadFunc,
	run func(context.Context, *apiprogram.Module) (map[string]*apivalues.Value, *apilang.RunSummary, error),
) lang.LoadFunc {
	return RejectCycles(func(ctx context.Context, path *apiprogram.Path) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
		if path.String() == "$inmem:second" {
			if len(mods) < 2 {
				return nil, nil, errors.New("not found")
			}

			return run(ctx, mods[1])
		}

		for _, mod := range mods {
			if path.Equal(mod.SourcePath()) {
				return run(ctx, mod)
			}
		}

		if load != nil {
			return load(ctx, path)
		}

		return nil, nil, errors.New("not found")
	})
}

func RejectCycles(load lang.LoadFunc) lang.LoadFunc {
	loads := make(map[string]bool)

	return func(ctx context.Context, path *apiprogram.Path) (map[string]*apivalues.Value, *apilang.RunSummary, error) {
		p := path.String()

		if finished, ok := loads[p]; ok && !finished {
			return nil, nil, errors.New("load cycle")
		}

		loads[p] = false

		vs, sum, err := load(ctx, path)

		loads[p] = true

		return vs, sum, err
	}
}
