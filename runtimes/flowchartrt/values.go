package flowchartrt

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (r *run) nodeToValue(path string, n *ast.Node) sdktypes.Value {
	desc := kittehs.Must1(sdktypes.ModuleFunctionFromProto(&sdktypes.ModuleFunctionPB{
		Input: []*sdktypes.ModuleFunctionFieldPB{{
			Name:  "**kwargs",
			Kwarg: true,
		}},
	}))

	return kittehs.Must1(sdktypes.NewFunctionValue(
		r.xid,
		n.Name,
		[]byte(path),
		[]sdktypes.FunctionFlag{},
		desc,
	))
}

func (r *run) valueToNode(v sdktypes.Value) (*ast.Node, error) {
	if !v.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("value is not valid")
	}

	f := v.GetFunction()
	if !f.IsValid() {
		return nil, sdkerrors.NewInvalidArgumentError("value is not a function")
	}

	if f.ExecutorID() != r.xid {
		return nil, nil
	}

	mod := r.modules[string(f.Data())]
	if mod == nil {
		return nil, nil
	}

	if n := mod.flowchart.GetNode(f.Name().String()); n != nil {
		return n, nil
	}

	return nil, fmt.Errorf("node %q: %w", f.Name(), sdkerrors.ErrNotFound)
}
