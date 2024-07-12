package flowchartrt

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/flosch/pongo2/v6"
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

	v, err := th.evalCELExpr(ctx, expr, true, nil)
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

func (th *thread) evalValue(ctx context.Context, u any, static bool, inputs map[string]any) (sdktypes.Value, error) {
	v, err := sdktypes.WrapValue(u)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return th.resolveValue(ctx, v, static, inputs)
}

// If static, not evaluated during runtime, else evaluated during runtime.
// Note that this is doing parsing and running. In the future, we'll separate these, so parsing
// is always be done at compile time, and evaluation might be running during compile type or runtime -
// depends on the case.
func (th *thread) evalCELExpr(ctx context.Context, expr string, static bool, inputs map[string]any) (sdktypes.Value, error) {
	prg, err := eval.Build(expr, static)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	if inputs == nil {
		if inputs, err = th.evalInputs(static); err != nil {
			return sdktypes.InvalidValue, err
		}
	}

	return th.evalProgram(ctx, prg, inputs)
}

func (th *thread) evalPongo2Expr(expr string, inputs map[string]any) (sdktypes.Value, error) {
	tpl, err := pongo2.FromString(expr)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	out, err := tpl.Execute(inputs)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewStringValue(out), nil
}

func (th *thread) evalGoExpr(expr string, inputs map[string]any) (sdktypes.Value, error) {
	t, err := template.New("").Parse(expr)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	var b strings.Builder

	if err := t.Execute(&b, inputs); err != nil {
		return sdktypes.InvalidValue, err
	}

	return sdktypes.NewStringValue(b.String()), err
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

func (th *thread) evalProgram(ctx context.Context, prg cel.Program, inputs map[string]any) (sdktypes.Value, error) {
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

func (th *thread) resolveValue(ctx context.Context, v sdktypes.Value, static bool, inputs map[string]any) (sdktypes.Value, error) {
	var err error
	if inputs == nil {
		if inputs, err = th.evalInputs(static); err != nil {
			return sdktypes.InvalidValue, err
		}
	}

	opts := th.frame().node.mod.flowchart.SafeOptions()

	return v.Map(func(v sdktypes.Value, mi *sdktypes.MapInfo) (sdktypes.Value, error) {
		if (mi.Kind == sdktypes.MapKindDictItemKey) || !v.IsString() {
			return v, nil
		}

		s := v.GetString().Value()

		interpolate := strings.HasPrefix(s, "^")

		if !interpolate && opts.InterpolateAllStrings {
			s = "^^" + s
			interpolate = true
		}

		if interpolate {
			kind, rest, ok := strings.Cut(s[1:], "^")
			if !ok {
				return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("invalid interpolation type")
			}

			if kind == "" {
				kind = opts.DefaultInterpolation
			}

			switch kind {
			case "cel", "":
				v, err = th.evalCELExpr(ctx, rest, static, inputs)

			case "p":
				v, err = th.evalPongo2Expr(rest, inputs)

			case "go":
				v, err = th.evalGoExpr(rest, inputs)

			default:
				v, err = sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("invalid interpolation type")
			}

			if err != nil {
				return sdktypes.InvalidValue, fmt.Errorf("interpolation %q: %w", s, err)
			}

			return v, sdktypes.ErrMapSkip
		} else if strings.HasPrefix(s, `\^`) {
			s = s[1:]
		}

		return sdktypes.NewStringValue(s), nil
	})
}
