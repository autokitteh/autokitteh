package aws

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/iancoleman/strcase"

	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func importServiceMethods(vars sdkservices.Vars, moduleName string, connect any) ([]sdkmodule.Optfn, error) {
	connectv, connectt := reflect.ValueOf(connect), reflect.TypeOf(connect)
	if connectt.NumOut() != 1 {
		return nil, sdklogger.DPanicOrReturn("connect method must return only the client")
	}

	client := connectt.Out(0)
	if client.Kind() != reflect.Ptr || client.Elem().Kind() != reflect.Struct {
		return nil, sdklogger.DPanicOrReturn("client is not a pointer to a struct")
	}

	opts := make([]sdkmodule.Optfn, 0, client.NumMethod()+1)

	for mi := range client.NumMethod() {
		m := client.Method(mi)

		// Ignore this method since it's non-standard and might reveal secrets.
		if m.Name == "Options" {
			continue
		}

		methodName, mt := strcase.ToSnake(m.Name), m.Type

		// Expecting self, context, params, optFns.
		if mt.NumIn() != 4 {
			return nil, sdklogger.DPanicOrReturn(fmt.Errorf("method %s.%q numin %d != 4", moduleName, methodName, mt.NumIn()))
		}

		pt := mt.In(2)
		if pt.Kind() != reflect.Ptr || pt.Elem().Kind() != reflect.Struct {
			return nil, sdklogger.DPanicOrReturn(fmt.Errorf("method %s.%q param invalid type: %v", moduleName, methodName, pt))
		}

		f := func(
			ctx context.Context,
			args []sdktypes.Value,
			kwargs map[string]sdktypes.Value,
		) (sdktypes.Value, error) {
			paramsValue := reflect.New(pt.Elem())

			if err := sdkmodule.UnpackArgs(args, kwargs, "params", paramsValue.Interface()); err != nil {
				return sdktypes.InvalidValue, err
			}

			cfg, err := getAWSConfig(ctx, vars)
			if err != nil {
				return sdktypes.InvalidValue, fmt.Errorf("token: %w", err)
			}

			if cfg == nil {
				return sdktypes.InvalidValue, errors.New("no config specified")
			}

			connectrets := connectv.Call([]reflect.Value{reflect.ValueOf(*cfg)})
			if len(connectrets) != 1 {
				return sdktypes.InvalidValue, errors.New("new client returned invalid values")
			}

			method := connectrets[0].MethodByName(m.Name) // must be original name.

			retvs := method.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				paramsValue,
			})

			if len(retvs) != 2 {
				return sdktypes.InvalidValue, fmt.Errorf("call returned %d values != expected 2", len(retvs))
			}

			outv, errv := retvs[0], retvs[1]

			if !errv.IsNil() {
				if err, ok := errv.Interface().(error); ok {
					return sdktypes.InvalidValue, err
				}

				return sdktypes.InvalidValue, errors.New("invalid error return")
			}

			out, err := sdktypes.WrapValue(outv.Interface())
			if err != nil {
				return sdktypes.InvalidValue, fmt.Errorf("return value conversion error: %w", err)
			}

			return out, nil
		}

		opts = append(opts, sdkmodule.ExportFunction(
			fmt.Sprintf("%s_%s", moduleName, methodName),
			f,
			sdkmodule.WithFuncDoc(fmt.Sprintf("%s.%s", moduleName, methodName)),
			sdkmodule.WithArg("params"),
		))
	}

	return opts, nil
}
