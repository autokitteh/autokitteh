package flowchartrt

import (
	"bytes"
	"context"
	"fmt"
	"maps"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type run struct {
	xid      sdktypes.ExecutorID
	compiled map[string][]byte
	cbs      *sdkservices.RunCallbacks
	globals  map[string]sdktypes.Value
	exports  map[string]sdktypes.Value
	modules  map[string]*module
	w        sdktypes.ValueWrapper
}

func (r *run) ID() sdktypes.RunID                 { return r.xid.ToRunID() }
func (r *run) ExecutorIDs() []sdktypes.ExecutorID { return []sdktypes.ExecutorID{r.xid} }
func (r *run) Values() map[string]sdktypes.Value  { return r.exports }
func (r *run) Close()                             {}

func (rt) Run(
	ctx context.Context,
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
	globals map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	r := run{
		xid:      sdktypes.NewExecutorID(runID),
		compiled: compiled,
		cbs:      cbs,
		globals:  globals,
		modules:  make(map[string]*module),
		w:        sdktypes.ValueWrapper{},
	}

	if err := r.loadAllModules(ctx); err != nil {
		return nil, err
	}

	mod := r.modules[mainPath]
	if mod == nil {
		return nil, fmt.Errorf("module %q not found", sdkerrors.ErrNotFound)
	}

	r.exports = mod.exports

	return &r, nil
}

func (r *run) loadAllModules(ctx context.Context) error {
	mods := make(map[string]*module, len(r.compiled))

	// Load all modules.
	for path, data := range r.compiled {
		f, err := ast.Read(path, bytes.NewBuffer(data))
		if err != nil {
			return fmt.Errorf("invalid compiled program: %w", err)
		}

		if mods[path], err = r.newModule(path, f); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
	}

	// Resolve load statements.
	for _, mod := range mods {
		for _, l := range mod.flowchart.Imports {
			var vs map[string]sdktypes.Value

			if isFlowchartPath(l.Path) {
				other := mods[l.Path]
				if other == nil {
					return fmt.Errorf("module %q: %w", l.Path, sdkerrors.ErrNotFound)
				}

				vs = kittehs.ListToMap(other.flowchart.Nodes, func(n *ast.Node) (name string, v sdktypes.Value) {
					return n.Name, r.nodeToValue(l.Path, n)
				})

				uwvs, err := kittehs.TransformMapValuesError(other.exports, sdktypes.WrapValue)
				if err != nil {
					return err
				}

				maps.Copy(vs, kittehs.TransformMapKeys(uwvs, kittehs.ToString))
			} else {
				var err error
				if vs, err = r.cbs.Load(ctx, r.xid.ToRunID(), l.Path); err != nil {
					return fmt.Errorf("load %q: %w", l.Path, err)
				}
			}

			mod.loads[l.Name] = vs
		}
	}

	r.modules = mods

	return nil
}

func (r *run) Call(ctx context.Context, entrypoint sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(args) > 0 {
		return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("positional args are not supported")
	}

	th, err := r.newThread(entrypoint, kwargs)
	if err != nil {
		return sdktypes.InvalidValue, th.newRuntimeError(err)
	}

	v, err := th.run(ctx)
	if err != nil {
		return sdktypes.InvalidValue, th.newRuntimeError(err)
	}

	return v, nil
}
