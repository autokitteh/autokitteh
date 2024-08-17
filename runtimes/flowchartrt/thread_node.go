package flowchartrt

import (
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type threadNode struct {
	th    *thread
	mod   *module
	node  *ast.Node
	value sdktypes.Value
}

func (tn *threadNode) toValue() sdktypes.Value { return tn.th.r.nodeToValue(tn.mod.path, tn.node) }

func (tn *threadNode) setValue(v sdktypes.Value) { tn.value = v }
