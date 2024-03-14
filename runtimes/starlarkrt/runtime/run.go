package runtime

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarktest"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/bootstrap"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/libs"
	"go.autokitteh.dev/autokitteh/runtimes/starlarkrt/internal/values"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	tlsKey = "autokitteh"

	// TODO: https://linear.app/autokitteh/issue/ENG-241/decide-how-to-fix-pr-250
	// testsGlobalName = "tests"
)

type tlsContext struct {
	goCtx   context.Context
	runID   sdktypes.RunID
	cbs     *sdkservices.RunCallbacks
	vctx    *values.Context
	globals starlark.StringDict
}

type run struct {
	runID    sdktypes.RunID
	compiled map[string][]byte

	vctx *values.Context
	cbs  *sdkservices.RunCallbacks

	// globals returns from initial evaluation of the script.
	globals starlark.StringDict

	// same as globals, but as autokitteh Values.
	exports map[string]sdktypes.Value
}

func (r *run) ID() sdktypes.RunID { return r.runID }

func Run(
	ctx context.Context,
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
	givenValues map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	prog, err := getProgram(compiled, mainPath)
	if err != nil {
		return nil, fmt.Errorf("invalid compiled program: %w", err)
	}
	if prog == nil {
		return nil, fmt.Errorf("not found: %q", mainPath)
	}

	vctx := &values.Context{Call: cbs.SafeCall, RunID: runID}

	predeclared, err := kittehs.TransformMapValuesError(givenValues, vctx.ToStarlarkValue)
	if err != nil {
		return nil, fmt.Errorf("converting values to starlark: %w", err)
	}

	// preload all  modules
	libs := libs.LoadModules(int64(kittehs.HashString64(runID.String())))

	// order matters here - builtins are the least important. predeclared can override them.
	// then we have the starlib libs, which can override both.
	predeclared, _ = kittehs.JoinMaps(builtins, predeclared, libs)

	th := &starlark.Thread{
		Name:  mainPath,
		Print: func(_ *starlark.Thread, text string) { cbs.SafePrint(ctx, runID, text) },
		Load: func(th *starlark.Thread, path string) (starlark.StringDict, error) {
			path = filepath.Join(filepath.Dir(th.Name), path)

			prog, err := getProgram(compiled, path)
			if err != nil {
				return nil, err
			}

			if prog == nil {
				globals, err := cbs.SafeLoad(ctx, runID, path)
				if err != nil {
					return nil, err
				}

				if globals == nil {
					return nil, sdkerrors.ErrNotFound
				}

				return kittehs.TransformMapValuesError(globals, vctx.ToStarlarkValue)
			}

			return prog.Init(th, predeclared)
		},
	}

	th.SetLocal(tlsKey, &tlsContext{
		goCtx: ctx,
		runID: runID,
		cbs:   cbs,
		vctx:  vctx,
	})

	var errorReporter errorReporter
	starlarktest.SetReporter(th, &errorReporter)

	bootstrappedValues, err := bootstrap.Run(th, predeclared)
	if err != nil {
		return nil, translateError(err, nil)
	}

	bootstrappedValues = kittehs.FilterMapKeys(bootstrappedValues, func(s string) bool {
		return s[0] != '_'
	})

	// We treat the bootstrap exported values as predeclared values in the actual program.
	predeclared, _ = kittehs.JoinMaps(predeclared, bootstrappedValues)

	globals, err := prog.Init(th, predeclared)
	if err != nil {
		// TODO: multierror.
		return nil, translateError(err, nil)
	}

	if len(errorReporter.errs) > 0 {
		// TODO(ENG-196): make program errors.
		return nil, fmt.Errorf("%s", errorReporter.Report())
	}

	globals = kittehs.FilterMapKeys(globals, func(s string) bool {
		return !strings.HasPrefix(s, "_")
	})

	// TODO: https://linear.app/autokitteh/issue/ENG-241/decide-how-to-fix-pr-250
	// if _, ok := globals[testsGlobalName]; !ok {
	// 	globals[testsGlobalName] = predeclared[testsGlobalName]
	// }

	exports, err := kittehs.TransformMapValuesError(globals, vctx.FromStarlarkValue)
	if err != nil {
		return nil, fmt.Errorf("converting values from starlark: %w", err)
	}

	return &run{runID: runID, compiled: compiled, exports: exports, globals: globals, vctx: vctx, cbs: cbs}, nil
}

func (r *run) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	fv, err := r.vctx.ToStarlarkValue(v)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	th := &starlark.Thread{
		Name:  v.GetFunction().UniqueID(),
		Print: func(_ *starlark.Thread, text string) { r.cbs.SafePrint(ctx, r.runID, text) },
	}

	th.SetLocal(tlsKey, &tlsContext{
		goCtx:   ctx,
		runID:   r.runID,
		globals: r.globals,
		cbs:     r.cbs,
		vctx:    r.vctx,
	})

	var errorReporter errorReporter
	starlarktest.SetReporter(th, &errorReporter)

	slargs, err := kittehs.TransformError(args, r.vctx.ToStarlarkValue)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("args transform: %w", err)
	}

	slkwargs := make([]starlark.Tuple, 0, len(kwargs))
	for k, v := range kwargs {
		slv, err := r.vctx.ToStarlarkValue(v)
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("kwarg with key %q transform: %w", k, err)
		}

		slkwargs = append(slkwargs, starlark.Tuple([]starlark.Value{
			starlark.String(k),
			slv,
		}))
	}

	slretv, err := starlark.Call(th, fv, slargs, slkwargs)
	if err != nil {
		return sdktypes.InvalidValue, translateError(err, nil)
	}

	if len(errorReporter.errs) > 0 {
		// TODO(ENG-196): make program errors.
		return sdktypes.InvalidValue, fmt.Errorf("%s", errorReporter.Report())
	}

	retv, err := r.vctx.FromStarlarkValue(slretv)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("return value transform: %w", err)
	}

	return retv, nil
}

func (r *run) Values() map[string]sdktypes.Value { return r.exports }

func (r *run) ExecutorID() sdktypes.ExecutorID { return sdktypes.NewExecutorID(r.runID) }

func (r *run) Close() {}

func getProgram(compiled map[string][]byte, path string) (*starlark.Program, error) {
	data := compiled[path]
	if data == nil {
		return nil, nil
	}

	prog, err := starlark.CompiledProgram(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("invalid compiled program: %w", err)
	}

	return prog, nil
}
