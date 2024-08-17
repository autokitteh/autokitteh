package flowchartrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (th *thread) runPrintAction(ctx context.Context, p *ast.PrintAction) error {
	v, err := th.evalValue(ctx, p.Value, false, nil)
	if err != nil {
		return fmt.Errorf("value: %w", err)
	}

	// First try to convert using the value prescribed manner...
	s, err := v.ToString()
	if err != nil {
		if !errors.Is(err, sdkerrors.ErrInvalidArgument{}) {
			return err
		}

		// ... then try to unwrap ...
		u, err := th.w.Unwrap(v)
		if err == nil {
			// ... and then marshal it using json to emit a human readable string.
			j, err := json.Marshal(u)
			if err == nil {
				s = string(j)
			} else {
				// it's not marshalable to json, to just emit what go things it should look like.
				s = fmt.Sprintf("%v", u)
			}
		} else {
			// ... not unwrappable! just emit the go string representation of the underlying proto.
			s = v.String()
		}
	}

	th.r.cbs.Print(ctx, th.r.xid.ToRunID(), s)

	return nil
}

func (th *thread) runSwitchAction(ctx context.Context, s ast.SwitchAction) (*ast.Node, error) {
	for _, c := range s {
		if c.If != nil {
			v, err := th.evalValue(ctx, c.If, false, nil)
			if err != nil {
				return nil, fmt.Errorf("condition: %w", err)
			}

			if !v.IsTruthy() {
				continue
			}
		}

		return th.evalNodeExpr(ctx, c.Goto)
	}

	return nil, nil
}

func (th *thread) runCallAction(ctx context.Context, c *ast.CallAction, setResult func(sdktypes.Value)) (*ast.Node, error) {
	if setResult == nil {
		setResult = func(sdktypes.Value) {}
	}

	if state := th.frame().getState("call"); len(state) > 0 {
		// this is a return from a node call. should always happen after a frame pop.
		th.setState("call", nil)
		setResult(th.frame().lastResult)
		return nil, nil
	}

	inputs, err := th.evalInputs(false)
	if err != nil {
		return nil, err
	}

	kwargs, err := kittehs.TransformMapValuesError(c.Args, func(v any) (sdktypes.Value, error) {
		return th.evalValue(ctx, v, false, inputs)
	})
	if err != nil {
		return nil, err
	}

	// TODO: evaluate at load time.
	v, err := th.evalCELExpr(ctx, c.Target, true, inputs)
	if err != nil {
		return nil, fmt.Errorf("target: %w", err)
	}

	if !v.IsFunction() {
		return nil, sdkerrors.NewInvalidArgumentError("target is niether a node or function")
	}

	if n, err := th.r.valueToNode(v); err != nil {
		return nil, err
	} else if n != nil {
		if c.Timeout != nil {
			return nil, sdkerrors.NewInvalidArgumentError("timeout not supported for node calls")
		}

		if c.Async {
			tn := th.findNode(n)
			if tn == nil {
				return nil, sdkerrors.NewInvalidArgumentError("node not found")
			}

			kwargs = map[string]sdktypes.Value{
				"inputs": sdktypes.NewDictValueFromStringMap(kwargs),
				"loc":    sdktypes.NewStringValuef("%s:%s", tn.mod.path, tn.node.Name),
			}

			// TODO: make this a first class citizen at the runtime callback interface.
			syscall, err := th.r.globals["ak"].GetKey(sdktypes.NewStringValue("syscall"))
			if err != nil {
				return nil, err
			}

			result, err := th.r.cbs.Call(ctx, th.r.xid.ToRunID(), syscall, []sdktypes.Value{sdktypes.NewStringValue("start")}, kwargs)
			if err != nil {
				return nil, err
			}

			setResult(result)

			return nil, nil
		}

		th.setState("call", map[string]sdktypes.Value{
			"args": sdktypes.NewDictValueFromStringMap(kwargs),
		})

		th.push(kwargs)

		return n, nil
	}

	if c.Async {
		return nil, sdkerrors.NewInvalidArgumentError("async not supported for non-node calls")
	}

	if c.Timeout != nil {
		tmov, err := th.evalValue(ctx, c.Timeout, false, inputs)
		if err != nil {
			return nil, fmt.Errorf("timeout: %w", err)
		}

		tmo, err := tmov.ToDuration()
		if err != nil {
			return nil, fmt.Errorf("timeout: %w", err)
		}

		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, tmo)
		defer cancel()
	}

	result, err := th.r.cbs.Call(ctx, th.r.xid.ToRunID(), v, nil, kwargs)
	if err != nil {
		if c.OnTimeoutGoto != "" && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return th.evalNodeExpr(ctx, c.OnTimeoutGoto)
		}

		return nil, err
	}

	setResult(result)

	return nil, nil
}

func (th *thread) runForEachAction(ctx context.Context, l *ast.ForEachAction) (*ast.Node, error) {
	setResult := func(v sdktypes.Value) {
		th.frame().updateResult(func(curr sdktypes.Value) sdktypes.Value {
			return kittehs.Must1(curr.Append(v))
		})
	}

	if state := th.frame().getState("call"); len(state) > 0 {
		// reset call state - if we end up here, we either going to start a new call
		// or returning from a call, since CallAction is embedded into this action.

		next, err := th.runCallAction(ctx, l.Call, setResult)
		if err != nil {
			return nil, err
		}

		if next != nil {
			return nil, sdkerrors.NewInvalidArgumentError("unexpected next node")
		}
	}

	state := th.frame().getState("foreach")
	if len(state) == 0 {
		v, err := th.evalValue(ctx, l.Items, false, nil)
		if err != nil {
			return nil, fmt.Errorf("items: %w", err)
		}

		state = map[string]sdktypes.Value{
			"items": v,
			"index": sdktypes.NewIntegerValue(-1),
		}
	}

	iv := state["index"]
	if !iv.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("no state index value")
	}

	var i int
	if err := iv.UnwrapInto(&i); err != nil {
		return nil, fmt.Errorf("state index: %w", err)
	}

	i++
	state["index"] = sdktypes.NewIntegerValue(i)

	itemsv := state["items"]
	if !itemsv.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("no state items value")
	}

	n, err := itemsv.Len()
	if err != nil {
		return nil, fmt.Errorf("items len: %w", err)
	}

	if i >= n {
		th.setState("foreach", nil)
		return nil, nil
	}

	if state["item"], err = itemsv.Index(i); err != nil {
		return nil, fmt.Errorf("items index: %w", err)
	}

	th.setState("foreach", state)

	th.frame().node.setValue(sdktypes.NewDictValueFromStringMap(state))

	next, err := th.runCallAction(ctx, l.Call, setResult)
	if err != nil {
		return nil, err
	}

	if next == nil {
		// call returned immiediately (non-node call), continue with the next item.
		next = th.frame().node.node
	}

	return next, nil
}
