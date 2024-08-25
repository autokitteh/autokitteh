package remotert

import (
	"fmt"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func unwrapMap(m map[string]sdktypes.Value) (map[string]any, error) {
	out := make(map[string]any, len(m))
	for key, val := range m {
		gv, err := unwrap(val)
		if err != nil {
			return nil, err
		}
		out[key] = gv
	}

	return out, nil
}

func unwrapList(values []sdktypes.Value) ([]any, error) {
	out := make([]any, 0, len(values))
	for _, val := range values {
		gv, err := unwrap(val)
		if err != nil {
			return nil, err
		}
		out = append(out, gv)
	}

	return out, nil
}

// We can't val.Unwrap since it fails on function types
func unwrap(val sdktypes.Value) (any, error) {
	switch {
	case val.IsBoolean():
		return val.GetBoolean().Value(), nil
	case val.IsBytes():
		return val.GetBytes().Value(), nil
	case val.IsDict():
		items, err := val.GetDict().ToStringValuesMap()
		if err != nil {
			return nil, err
		}

		return unwrapMap(items)
	case val.IsDuration():
		return val.GetDuration().Value(), nil
	case val.IsFloat():
		return val.GetFloat().Value(), nil
	case val.IsFunction():
		fnName := val.GetFunction().Name().String()
		return fmt.Sprintf("func:%s", fnName), nil
	case val.IsInteger():
		return val.GetInteger().Value(), nil
	case val.IsList():
		values := val.GetList().Values()
		return unwrapList(values)
	case val.IsModule():
		members := val.GetModule().Members()
		return unwrapMap(members)
	case val.IsNothing():
		return nil, nil
	case val.IsSet():
		values := val.GetSet().Values()
		return unwrapList(values)
	case val.IsString():
		return val.GetString().Value(), nil
	case val.IsStruct():
		fields := val.GetStruct().Fields()
		return unwrapMap(fields)
	case val.IsSymbol():
		return val.GetSymbol().String(), nil
	case val.IsTime():
		return val.GetTime().Value(), nil
	}

	return nil, fmt.Errorf("unknown type: %#v", val)
}
