package sdkruntimes

import (
	"context"
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type RunParams struct {
	Runtimes             sdkservices.Runtimes
	BuildFile            *sdkbuildfile.BuildFile
	Globals              map[string]sdktypes.Value
	RunID                sdktypes.RunID
	FallthroughCallbacks sdkservices.RunCallbacks
	EntryPointPath       string
}

// Run executes a build file and manages it across multiple runtimes.
// fallthourghCallbacks.{Load,Call} are called only for dynamic modules
// (modules that are supplied from integrations).
func Run(ctx context.Context, params RunParams) (sdkservices.Run, error) {
	group := &group{
		mainID: params.RunID,
		runs:   make(map[string]sdkservices.Run),
	}

	cbs := sdkservices.RunCallbacks{
		Print: params.FallthroughCallbacks.SafePrint,
		Call: func(ctx context.Context, runID sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
			if !sdktypes.FunctionValueHasExecutorID(v) ||
				group.runs[sdktypes.GetFunctionValueExecutorID(v).ToRunID().String()] == nil {
				return params.FallthroughCallbacks.SafeCall(ctx, runID, v, args, kwargs)
			}

			return group.Call(ctx, v, args, kwargs)
		},
		NewRunID: params.FallthroughCallbacks.NewRunID,
	}

	cache := make(map[string]map[string]sdktypes.Value)

	cbs.Load = func(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
		exports, ok := cache[path]
		if ok {
			return exports, nil
		}

		loadRunID := params.FallthroughCallbacks.SafeNewRunID()

		runParams := params
		runParams.Globals = nil // TODO: globals: figure out which values to pass here.
		runParams.RunID = loadRunID
		runParams.FallthroughCallbacks = cbs

		r, err := run(ctx, runParams, path)
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				return params.FallthroughCallbacks.SafeLoad(ctx, rid, path)
			}

			return nil, err
		}

		group.runs[loadRunID.String()] = r

		cache[path] = r.Values()

		return r.Values(), err
	}

	runParams := params
	runParams.FallthroughCallbacks = cbs

	r, err := run(ctx, runParams, params.EntryPointPath)
	if r != nil {
		r.Close()
	}

	group.runs[group.mainID.String()] = r

	return group, err
}

func run(ctx context.Context, params RunParams, path string) (sdkservices.Run, error) {
	ls, err := params.Runtimes.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list runtimes: %w", err)
	}

	rtd := MatchRuntimeByPath(ls, path)
	if rtd == nil {
		return nil, sdkerrors.ErrNotFound
	}

	found := -1

	for i, brt := range params.BuildFile.Runtimes {
		if brt.Info.Name.String() == sdktypes.GetRuntimeName(rtd).String() {
			found = i
			break
		}
	}

	if found < 0 {
		return nil, fmt.Errorf("no matching runtime for path %q", path)
	}

	brt := params.BuildFile.Runtimes[found]

	rt, err := params.Runtimes.New(ctx, brt.Info.Name)
	if err != nil {
		return nil, fmt.Errorf("new runtime: %w", err)
	}

	return rt.Run(
		ctx,
		params.RunID,
		path,
		sdktypes.GetBuildArtifactCompiledData(brt.Artifact),
		params.Globals,
		&params.FallthroughCallbacks,
	)
}
