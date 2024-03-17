package grpc

import (
	"context"
	"errors"
	"sort"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkexecutor"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var safeForJsonWrapper = sdktypes.ValueWrapper{SafeForJSON: true}

func parsePayload(args []sdktypes.Value, kwargs map[string]sdktypes.Value) (map[string]any, error) {
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

func createGRPCCallWrapper(functionName string) sdkexecutor.Function {
	return func(ctx context.Context, v []sdktypes.Value, m map[string]sdktypes.Value) (sdktypes.Value, error) {
		payload, err := parsePayload(v, m)
		if err != nil {
			return sdktypes.Nothing, err
		}

		hostport := string(sdkmodule.FunctionDataFromContext(ctx))

		conn, err := grpc.Dial(hostport, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return sdktypes.Nothing, err
		}
		defer conn.Close()

		s, err := newGRPCClient(conn)
		if err != nil {
			return sdktypes.Nothing, err
		}

		res, err := s.invoke(functionName, payload)
		if err != nil {
			return sdktypes.Nothing, err
		}

		return sdktypes.DefaultValueWrapper.Wrap(res)
	}
}

func newGRPCModule(config string) ([]sdkmodule.Optfn, error) {
	addr := string(config)

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	s, err := newGRPCClient(conn)
	if err != nil {
		return nil, err
	}

	svcs, err := s.descSource.ListServices()
	if err != nil {
		return nil, err
	}

	var fns []method
	for _, svc := range svcs {
		methods, err := s.listMethods(svc)
		if err != nil {
			return nil, err
		}
		fns = append(fns, methods...)
	}

	sort.SliceStable(fns, func(i, j int) bool {
		return fns[i].Name < fns[j].Name
	})

	opts := []sdkmodule.Optfn{
		sdkmodule.WithConfigAsData(),
	}

	for _, f := range fns {
		opts = append(opts, kittehs.Transform(f.Constants, func(c string) sdkmodule.Optfn {
			return sdkmodule.ExportValue(c, sdkmodule.WithValue(kittehs.Must1(sdktypes.WrapValue(c))))
		})...)
	}

	opts = append(opts, kittehs.Transform(fns, func(f method) sdkmodule.Optfn {
		return sdkmodule.ExportFunction(f.Name, createGRPCCallWrapper(f.Fullname), sdkmodule.WithArgs(f.Inputs...))
	})...)

	return opts, err
}
