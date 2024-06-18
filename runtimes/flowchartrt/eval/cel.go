package eval

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

var (
	staticEvalOpts = []cel.EnvOption{
		cel.Variable("globals", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("imports", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("nodes", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("values", cel.MapType(cel.StringType, cel.AnyType)),
		ext.Bindings(),
		ext.Encoders(),
		ext.Lists(),
		ext.Math(),
		ext.NativeTypes(),
		ext.Sets(),
		ext.Strings(),
	}

	staticEvalEnv = kittehs.Must1(cel.NewEnv(staticEvalOpts...))

	dynamicEvalOpts = append(
		staticEvalOpts,
		cel.Variable("args", cel.MapType(cel.StringType, cel.AnyType)),
		cel.Variable("states", cel.MapType(cel.StringType, cel.AnyType)),
	)

	dynamicEvalEnv = kittehs.Must1(cel.NewEnv(dynamicEvalOpts...))
)
