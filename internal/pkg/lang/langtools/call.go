package langtools

import (
	"context"
	"fmt"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apilang"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apivalues"
	"github.com/autokitteh/autokitteh/internal/pkg/lang"
)

func CallFunction(
	ctx context.Context,
	cat lang.Catalog,
	env *lang.RunEnv,
	v *apivalues.Value,
	args []*apivalues.Value,
	kws map[string]*apivalues.Value,
) (lang.Lang, *apivalues.Value, *apilang.RunSummary, error) {
	fn, ok := v.Get().(apivalues.FunctionValue)
	if !ok {
		return nil, nil, nil, fmt.Errorf("value is not a function")
	}

	l, err := cat.Acquire(fn.Lang, env.Scope)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("lang %q: %w", fn.Lang, err)
	}

	rv, sum, err := l.CallFunction(ctx, env, v, args, kws)

	return l, rv, sum, err
}
