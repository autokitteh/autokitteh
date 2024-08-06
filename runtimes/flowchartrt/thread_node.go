package flowchartrt

import (
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type threadNode struct {
	th     *thread
	mod    *module
	node   *ast.Node
	states map[string]sdktypes.Value
}

func (tn *threadNode) toValue() sdktypes.Value { return tn.th.r.nodeToValue(tn.mod.path, tn.node) }

func (tn *threadNode) setState(k string, v map[string]sdktypes.Value) {
	if v != nil {
		if tn.states == nil {
			tn.states = make(map[string]sdktypes.Value)
		}

		sym := sdktypes.NewSymbolValue(sdktypes.NewSymbol(k))

		tn.states[k] = kittehs.Must1(sdktypes.NewStructValue(sym, v))

		return
	}

	if tn.states != nil {
		delete(tn.states, k)
	}
}
