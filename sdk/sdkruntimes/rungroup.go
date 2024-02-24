package sdkruntimes

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// Run group. Manages all the sub-runs (loaded runtime paths).
type group struct {
	mainID sdktypes.RunID
	runs   map[string]sdkservices.Run
}

var _ sdkservices.Run = (*group)(nil)

func (g *group) ID() sdktypes.RunID { return g.mainID }

func (g *group) Values() map[string]sdktypes.Value {
	if r := g.runs[g.mainID.String()]; r != nil {
		return r.Values()
	}

	return nil
}

func (g *group) Close() {
	for _, r := range g.runs {
		r.Close()
	}
}

func (g *group) ExecutorID() sdktypes.ExecutorID { return sdktypes.NewExecutorID(g.mainID) }

func (g *group) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if !sdktypes.IsFunctionValue(v) {
		return nil, fmt.Errorf("callee must be a function value")
	}

	executorID := sdktypes.GetFunctionValueExecutorID(v)

	runID := executorID.ToRunID()
	if runID == nil {
		return nil, fmt.Errorf("executor is not a run")
	}

	run, ok := g.runs[runID.String()]
	if !ok {
		return nil, fmt.Errorf("run id not found: %w", sdkerrors.ErrNotFound)
	}

	return run.Call(ctx, v, args, kwargs)
}
