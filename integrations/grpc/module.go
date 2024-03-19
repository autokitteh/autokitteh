package grpc

import (
	"context"
	"errors"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var safeForJsonWrapper = sdktypes.ValueWrapper{SafeForJSON: true}

func parseArgs(args []sdktypes.Value, kwargs map[string]sdktypes.Value) (map[string]any, error) {
	if len(args) > 1 {
		return nil, errors.New("args len should be 0 or 1")
	}

	if len(args) == 1 {
		if len(kwargs) != 0 {
			return nil, errors.New("either provide one dict arg or kwargs")
		}

		if !args[0].IsDict() {
			return nil, errors.New("args has to be dict")
		}

		var result map[string]any
		if err := safeForJsonWrapper.UnwrapInto(&result, args[0]); err != nil {
			return nil, err
		}
		return result, nil
	}

	return kittehs.TransformMapError(kwargs, func(key string, val sdktypes.Value) (string, any, error) {
		d, err := safeForJsonWrapper.Unwrap(val)
		if err != nil {
			return "", nil, err
		}

		return key, d, nil
	})
}

func handleGenericGRPCCall() sdkexecutor.Function {
	return func(ctx context.Context, v []sdktypes.Value, m map[string]sdktypes.Value) (sdktypes.Value, error) {
		args, err := parseArgs(v, m)
		if err != nil {
			return sdktypes.Nothing, err
		}

		hostport, ok := args["host"].(string)
		if !ok {
			return sdktypes.Nothing, errors.New("host is required")
		}

		conn, err := grpc.Dial(hostport, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return sdktypes.Nothing, err
		}
		defer conn.Close()

		s, err := newGRPCClient(conn)
		if err != nil {
			return sdktypes.Nothing, err
		}

		service, ok := args["service"].(string)
		if !ok {
			return sdktypes.Nothing, errors.New("service is required")
		}
		method, ok := args["method"].(string)
		if !ok {
			return sdktypes.Nothing, errors.New("method is required")
		}

		payload := map[string]any{}
		if data, ok := args["payload"]; ok {
			if payload, ok = data.(map[string]any); !ok {
				return sdktypes.Nothing, errors.New("payload has to be dict")
			}
		}

		funcName := fmt.Sprintf("%s.%s", service, method)
		res, err := s.invoke(funcName, payload)
		if err != nil {
			return sdktypes.Nothing, err
		}

		return sdktypes.DefaultValueWrapper.Wrap(res)
	}
}
