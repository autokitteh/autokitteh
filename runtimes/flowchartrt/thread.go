package flowchartrt

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type thread struct {
	r         *run
	nodes     []*threadNode
	callstack []*frame // [0] is topmost (active).
	w         sdktypes.ValueWrapper
}

func (th *thread) frame() *frame { return th.callstack[0] }

func (th *thread) findNode(n *ast.Node) *threadNode {
	_, tn := kittehs.FindFirst(th.nodes, func(tn *threadNode) bool { return tn.node == n })
	return tn
}

func (r *run) newThread(entrypoint sdktypes.Value, args map[string]sdktypes.Value) (*thread, error) {
	rootNode, err := r.valueToNode(entrypoint)
	if err != nil {
		return nil, fmt.Errorf("entrypoint: %w", err)
	}

	th := thread{r: r}

	th.w = sdktypes.ValueWrapper{
		UnwrapFunction: func(v sdktypes.Value) (any, error) {
			// If a function is a const function, simplify it by converting it into a value.
			if fv := v.GetFunction(); fv.IsConst() {
				if v, err := fv.ConstValue(); err == nil {
					return th.w.Unwrap(v)
				}

				// if err != nil, function always return error. user wont always refer to it,
				// so just return it as a function to be lazily evaluated later by the user
				// if they wish to.
			}
			return v, nil
		},
		SafeForJSON: true,
	}

	var curr *threadNode

	for _, mod := range r.modules {
		for _, v := range mod.exports {
			if !v.IsFunction() {
				continue
			}

			node, err := r.valueToNode(v)
			if err != nil {
				return nil, err
			}

			thNode := &threadNode{th: &th, mod: mod, node: node}

			th.nodes = append(th.nodes, thNode)

			if node == rootNode {
				curr = thNode
			}
		}
	}

	if curr == nil {
		return nil, fmt.Errorf("node %v: %w", rootNode.Name, sdkerrors.ErrNotFound)
	}

	th.callstack = []*frame{{node: curr, args: args}}

	return &th, nil
}

func (th *thread) push(args map[string]sdktypes.Value) {
	f := &frame{
		node: th.frame().node,
		args: args,
	}

	// Frames inherits the results from previous frames. This allows them to use
	// results from nodes that were previously processed.
	if rs, ok := th.frame().states["results"]; ok {
		f.states = map[string]sdktypes.Value{"results": rs}
	}

	th.callstack = append([]*frame{f}, th.callstack...)
}

func (th *thread) pop(ctx context.Context) (next *ast.Node, result sdktypes.Value, err error) {
	frame := th.frame()

	if frame.node.node.Goto != "" {
		next, err = th.evalNodeExpr(ctx, frame.node.node.Goto)
		if err != nil {
			return nil, sdktypes.InvalidValue, fmt.Errorf("next: %w", err)
		}
	}

	th.callstack = th.callstack[1:]

	result = frame.getResult()

	return
}

func (th *thread) run(ctx context.Context) (sdktypes.Value, error) {
	result := sdktypes.Nothing

	for len(th.callstack) > 0 {
		next, err := th.step(ctx, th.frame())
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if next == nil {
			if next, result, err = th.pop(ctx); err != nil {
				return sdktypes.InvalidValue, fmt.Errorf("next: %w", err)
			}

			if !result.IsValid() {
				result = sdktypes.Nothing
			}

			if len(th.callstack) > 0 {
				th.frame().lastResult = result
			}

			if next == nil {
				continue
			}
		}

		curr := th.findNode(next)
		if curr == nil {
			return sdktypes.InvalidValue, fmt.Errorf("node %v: %w", next.Name, sdkerrors.ErrNotFound)
		}

		th.frame().node = curr
	}

	return result, nil
}

func (th *thread) debugCallstack() []sdktypes.CallFrame {
	callstack := make([]sdktypes.CallFrame, len(th.callstack))
	for i, f := range th.callstack {
		locals := map[string]sdktypes.Value{
			"states": sdktypes.NewDictValueFromStringMap(f.states),
			"args":   sdktypes.NewDictValueFromStringMap(f.args),
		}

		callstack[i] = sdktypes.NewCallFrame(
			f.node.node.Name,
			f.node.node.Location(),
			locals,
		)
	}

	return callstack
}

func (th *thread) debugTrace(ctx context.Context, reason string) {
	th.r.cbs.SafeDebugTrace(ctx, th.r.ID(), th.debugCallstack(), map[string]string{"reason": reason})
}

func (th *thread) step(ctx context.Context, frame *frame) (*ast.Node, error) {
	node := frame.node.node

	th.debugTrace(ctx, "prestep")

	if a := node.Action; a != nil {
		var (
			err  error
			next *ast.Node // if not nil, which node to go to next.
		)

		switch {
		case a.Print != nil:
			err = th.runPrintAction(ctx, a.Print)
		case a.Switch != nil:
			next, err = th.runSwitchAction(ctx, a.Switch)
		case a.Call != nil:
			next, err = th.runCallAction(ctx, a.Call, th.frame().setResult)
		case a.ForEach != nil:
			next, err = th.runForEachAction(ctx, a.ForEach)
		default:
			err = sdkerrors.ErrNotImplemented
		}

		if err != nil {
			return nil, err
		}

		if next != nil {
			return next, nil
		}
	}

	if node.Result != nil {
		result, err := th.evalValue(ctx, node.Result, false)
		if err != nil {
			return nil, fmt.Errorf("result: %w", err)
		}

		th.frame().setResult(result)
	}

	th.debugTrace(ctx, "poststep")

	// evaluate default next node.
	if next := node.Goto; next != "" {
		return th.evalNodeExpr(ctx, node.Goto)
	}

	return nil, nil
}

func (th *thread) newRuntimeError(err error) error {
	if sdktypes.IsProgramError(err) {
		return err
	}

	return sdktypes.NewProgramError(sdktypes.NewStringValue(err.Error()), th.debugCallstack(), nil).ToError()
}
