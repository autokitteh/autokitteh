package flowchartrt

import (
	"maps"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/flowchartrt/ast"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type module struct {
	r         *run
	path      string
	flowchart *ast.Flowchart
	loads     map[string]map[string]sdktypes.Value
	exports   map[string]sdktypes.Value
	values    map[string]sdktypes.Value
}

func (r *run) newModule(path string, f *ast.Flowchart) (*module, error) {
	vs := kittehs.ListToMap(f.Nodes, func(n *ast.Node) (name string, v sdktypes.Value) {
		return n.Name, r.nodeToValue(path, n)
	})

	us, err := kittehs.TransformMapValuesError(f.Values, sdktypes.WrapValue)
	if err != nil {
		return nil, err
	}

	maps.Copy(vs, kittehs.TransformMapKeys(us, kittehs.ToString))

	return &module{
		r:         r,
		path:      path,
		flowchart: f,
		loads:     make(map[string]map[string]sdktypes.Value),
		values:    vs,
		exports:   kittehs.FilterMapKeys(vs, func(n string) bool { return !strings.HasPrefix(n, "_") }),
	}, nil
}
