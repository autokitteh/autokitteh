package flowchartrt

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/google/cel-go/cel"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/eval"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (th *thread) evalNodeExpr(ctx context.Context, expr string) (*ast.Node, error) {
	if expr == "" {
		return nil, nil
	}

	v, err := th.evalExpr(ctx, expr, true)
	if err != nil {
		return nil, err
	}

	n, err := th.r.valueToNode(v)
	if err != nil {
		return nil, err
	}

	if n == nil {
		return nil, sdkerrors.ErrNotFound
	}

	return n, nil
}

func (th *thread) evalValue(ctx context.Context, v any, static bool) (sdktypes.Value, error) {
	if expr, ok := v.(string); ok {
		return th.evalExpr(ctx, expr, static)
	}

	return sdktypes.WrapValue(v)
}

// If static, not evaluated during runtime, else evaluated during runtime.
// Note that this is doing parsing and running. In the future, we'll separate these, so parsing
// is always be done at compile time, and evaluation might be running during compile type or runtime -
// depends on the case.
func (th *thread) evalExpr(ctx context.Context, expr string, static bool) (sdktypes.Value, error) {
	prg, err := eval.Build(expr, static)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return th.evalProgram(ctx, prg, static)
}

func (th *thread) evalInputs(static bool) (map[string]any, error) {
	nodes := kittehs.ListToMap(th.nodes, func(n *threadNode) (string, any) { return n.node.Name, n.toValue() })

	frame := th.frame()

	inputs := map[string]any{"nodes": nodes}

	if th.frame().node.mod == th.callstack[len(th.callstack)-1].node.mod {
		// allow globals access only from the root module.
		var err error
		if inputs["globals"], err = kittehs.TransformMapValuesError(th.r.globals, th.w.Unwrap); err != nil {
			return nil, fmt.Errorf("transform globals: %w", err)
		}
	}

	if !static {
		var err error
		if inputs["states"], err = kittehs.TransformMapValuesError(frame.states, th.w.Unwrap); err != nil {
			return nil, fmt.Errorf("transform state: %w", err)
		}
	}

	if curr := frame.node; curr != nil {
		var err error

		inputs["values"], err = kittehs.TransformMapValuesError(curr.mod.values, th.w.Unwrap)
		if err != nil {
			return nil, fmt.Errorf("transform values: %w", err)
		}

		inputs["imports"], err = kittehs.TransformMapValuesError(curr.mod.loads, func(vs map[string]sdktypes.Value) (map[string]any, error) {
			return kittehs.TransformMapValuesError(vs, th.w.Unwrap)
		})
		if err != nil {
			return nil, fmt.Errorf("transform imports: %w", err)
		}

		if !static {
			if inputs["args"], err = kittehs.TransformMapValuesError(frame.args, th.w.Unwrap); err != nil {
				return nil, fmt.Errorf("unwrap input: %w", err)
			}
		}
	}

	return inputs, nil
}

func (th *thread) evalProgram(ctx context.Context, prg cel.Program, static bool) (sdktypes.Value, error) {
	inputs, err := th.evalInputs(static)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	out, _, err := prg.ContextEval(ctx, inputs)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("eval: %w", err)
	}

	v, err := th.w.Wrap(out.Value())
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("wrap: %w", err)
	}

	return v, nil
}

func (th *thread) evalArg(ctx context.Context, v any, inputs map[string]any) (sdktypes.Value, error) {
	var (
		result sdktypes.Value
		err    error
	)

	if th.frame().node.mod.flowchart.HasPragma("argsnoexpr") {
		result, err = sdktypes.WrapValue(v)
	} else {
		result, err = th.evalValue(ctx, v, false)
	}

	if err != nil {
		return sdktypes.InvalidValue, err
	}

	result, err = result.Map(func(v sdktypes.Value) (sdktypes.Value, error) {
		if !v.IsString() {
			return v, nil
		}

		s := v.GetString().Value()

		t, err := template.New("").Parse(s)
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("invalid interpolation string %q: %w", s, err)
		}

		var b strings.Builder

		if err := t.Execute(&b, inputs); err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("interpolation %q: %w", s, err)
		}

		return sdktypes.NewStringValue(b.String()), err
	})

	return result, err
}

// Arguments also perform string interpolation on the output.
func (th *thread) evalArgs(ctx context.Context, vs map[string]any) (map[string]sdktypes.Value, error) {
	inputs, err := th.evalInputs(false)
	if err != nil {
		return nil, err
	}

	eval := func(v any) (sdktypes.Value, error) { return th.evalArg(ctx, v, inputs) }

	return kittehs.TransformMapValuesError(vs, eval)
}
