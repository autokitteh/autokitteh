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
	runs   map[sdktypes.ExecutorID]sdkservices.Run
}

var _ sdkservices.Run = (*group)(nil)

func (g *group) ID() sdktypes.RunID { return g.mainID }

func (g *group) Values() map[string]sdktypes.Value {
	if r := g.runs[sdktypes.NewExecutorID(g.mainID)]; r != nil {
		return r.Values()
	}

	return nil
}

func (g *group) Close() {
	for _, r := range g.runs {
		r.Close()
	}
}

func (g *group) ExecutorIDs() (xids []sdktypes.ExecutorID) {
	for _, r := range g.runs {
		xids = append(xids, r.ExecutorIDs()...)
	}

	return
}

func (g *group) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if !v.IsFunction() {
		return sdktypes.InvalidValue, fmt.Errorf("callee must be a function value")
	}

	executorID := v.GetFunction().ExecutorID()

	runID := executorID.ToRunID()
	if !runID.IsValid() {
		return sdktypes.InvalidValue, fmt.Errorf("executor is not a run")
	}

	run, ok := g.runs[sdktypes.NewExecutorID(runID)]
	if !ok {
		return sdktypes.InvalidValue, fmt.Errorf("run id not found: %w", sdkerrors.ErrNotFound)
	}

	return run.Call(ctx, v, args, kwargs)
}
