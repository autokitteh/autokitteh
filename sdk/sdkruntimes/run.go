package sdkruntimes

import (
	"context"
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type RunParams struct {
	Runtimes             sdkservices.Runtimes
	BuildFile            *sdkbuildfile.BuildFile
	Globals              map[string]sdktypes.Value
	RunID                sdktypes.RunID
	SessionID            sdktypes.SessionID
	FallthroughCallbacks sdkservices.RunCallbacks
	EntryPointPath       string
}

type runCallbacks struct {
	sdkservices.NopRunCallbacks
	params RunParams
	group  *group
	cache  map[string]map[string]sdktypes.Value
}

func (cbs runCallbacks) Call(ctx context.Context, rid sdktypes.RunID, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if xid := v.GetFunction().ExecutorID(); xid.IsValid() || cbs.group.runs[xid] == nil {
		return cbs.params.FallthroughCallbacks.Call(ctx, rid, v, args, kwargs)
	}

	return cbs.group.Call(ctx, v, args, kwargs)
}

func (cbs runCallbacks) Load(ctx context.Context, rid sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
	exports, ok := cbs.cache[path]
	if ok {
		return exports, nil
	}

	loadRunID := cbs.params.FallthroughCallbacks.NewRunID()

	runParams := cbs.params
	runParams.Globals = nil // TODO: globals: figure out which values to pass here.
	runParams.RunID = loadRunID
	runParams.FallthroughCallbacks = cbs
	runParams.SessionID = cbs.params.SessionID

	r, err := run(ctx, runParams, path)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return cbs.params.FallthroughCallbacks.Load(ctx, rid, path)
		}

		return nil, err
	}

	cbs.group.runs[sdktypes.NewExecutorID(loadRunID)] = r

	cbs.cache[path] = r.Values()

	return r.Values(), err
}

func (runCallbacks) NewRunID() sdktypes.RunID { return sdktypes.NewRunID() }

// Run executes a build file and manages it across multiple runtimes.
// fallthourghCallbacks.{Load,Call} are called only for dynamic modules
// (modules that are supplied from integrations).
func Run(ctx context.Context, params RunParams) (sdkservices.Run, error) {
	group := &group{
		mainID: params.RunID,
		runs:   make(map[sdktypes.ExecutorID]sdkservices.Run),
	}

	runParams := params
	runParams.FallthroughCallbacks = runCallbacks{
		params: params,
		group:  group,
		cache:  make(map[string]map[string]sdktypes.Value),
	}

	r, err := run(ctx, runParams, params.EntryPointPath)
	if r != nil {
		r.Close()
	}

	group.runs[sdktypes.NewExecutorID(group.mainID)] = r

	return group, err
}

func run(ctx context.Context, params RunParams, path string) (sdkservices.Run, error) {
	ls, err := params.Runtimes.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list runtimes: %w", err)
	}

	rtd, ok := MatchRuntimeByPath(ls, path)
	if !ok {
		return nil, sdkerrors.ErrNotFound
	}

	found := -1

	for i, brt := range params.BuildFile.Runtimes {
		if brt.Info.Name == rtd.Name() {
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
		params.SessionID,
		path,
		brt.Artifact.CompiledData(),
		params.Globals,
		params.FallthroughCallbacks,
	)
}
