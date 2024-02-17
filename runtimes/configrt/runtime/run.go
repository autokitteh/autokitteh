package runtime

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type run struct {
	runID   sdktypes.RunID
	exports map[string]sdktypes.Value
}

var _ sdkservices.Run = &run{}

func (r *run) ID() sdktypes.RunID                { return r.runID }
func (r *run) Values() map[string]sdktypes.Value { return r.exports }
func (r *run) ExecutorID() sdktypes.ExecutorID   { return sdktypes.NewExecutorID(r.runID) }

func (r *run) getFunc(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	var key string

	if err := sdkmodule.UnpackArgs(args, kwargs, "key", &key); err != nil {
		return nil, err
	}

	v, ok := r.exports[key]
	if !ok {
		return sdktypes.NewNothingValue(), nil
	}

	return v, nil
}

func (r *run) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if !sdktypes.IsFunctionValue(v) {
		return nil, sdkerrors.ErrInvalidArgument
	}

	name := sdktypes.GetFunctionValueName(v)

	switch name.String() {
	case "get":
		return r.getFunc(ctx, args, kwargs)
	default:
		return nil, fmt.Errorf("unrecognized function name %q", name)
	}
}

func (*run) Close() {}

func Run(
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
) (sdkservices.Run, error) {
	data, ok := compiled[mainPath]
	if !ok {
		return nil, fmt.Errorf("not in compiled data: %q", mainPath)
	}

	exports, err := evaluateBytes(data)
	if err != nil {
		return nil, err
	}

	return &run{exports: exports, runID: runID}, nil
}

func evaluateBytes(data []byte) (map[string]sdktypes.Value, error) {
	var pb sdktypes.ValuePB

	if err := proto.Unmarshal(data, &pb); err != nil {
		return nil, fmt.Errorf("invalid compiled data: %w", err)
	}

	v, err := sdktypes.ValueFromProto(&pb)
	if err != nil {
		return nil, fmt.Errorf("invalid value data: %w", err)
	}

	return evaluateValue(v)
}

func evaluateValue(v sdktypes.Value) (map[string]sdktypes.Value, error) {
	var exports map[string]sdktypes.Value

	switch sdktypes.GetValue(v).(type) {
	case sdktypes.ListValue:
		i := 0
		exports = kittehs.ListToMap(
			sdktypes.GetListValue(v),
			func(v sdktypes.Value) (string, sdktypes.Value) {
				i++
				return fmt.Sprintf("_%d", i), v
			},
		)
	case sdktypes.SetValue:
		i := 0
		exports = kittehs.ListToMap(
			sdktypes.GetListValue(v),
			func(v sdktypes.Value) (string, sdktypes.Value) {
				i++
				return fmt.Sprintf("_%d", i), v
			},
		)
	case sdktypes.DictValue:
		var err error
		exports, err = kittehs.ListToMapError(
			sdktypes.GetDictValue(v),
			func(it *sdktypes.DictValueItem) (string, sdktypes.Value, error) {
				if !sdktypes.IsStringValue(it.K) {
					return "", nil, fmt.Errorf("dict key is not a string")
				}
				return sdktypes.GetStringValue(it.K), it.V, nil
			},
		)
		if err != nil {
			return nil, err
		}
	case sdktypes.StructValue:
		_, exports = sdktypes.GetStructValue(v)
	case sdktypes.ModuleValue:
		_, exports = sdktypes.GetModuleValue(v)
	default:
		return nil, fmt.Errorf("unhandled value type")
	}

	return exports, nil
}
