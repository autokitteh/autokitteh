package eval

import (
	"github.com/google/cel-go/cel"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func Build(v any, static bool) (cel.Program, error) {
	expr, ok := v.(string)
	if !ok {
		// not an expr - just a value.
		return nil, nil
	}

	if expr == "" {
		return nil, sdkerrors.NewInvalidArgumentError("empty expression")
	}

	env := staticEvalEnv
	if !static {
		env = dynamicEvalEnv
	}

	// TODO: move to flowchart ast load?
	ast, issues := env.Compile(expr)
	if err := issues.Err(); err != nil {
		return nil, err
	}

	return env.Program(ast)
}
