package aws

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/iancoleman/strcase"

	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

func getAWSConfig(cfg []byte) (*aws.Config, error) {
	if string(cfg) == "default" {
		return defaultAWSConfig, nil
	}

	parts := strings.Split(string(cfg), ",")
	if len(parts) != 3 && len(parts) != 4 {
		return nil, errors.New("invalid config - exprecting \"region,access_key_id,secret_key,token\"")
	}

	var awsToken string
	if len(parts) == 4 {
		awsToken = parts[3]
	}

	awsCfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(parts[1], parts[2], awsToken)),
		config.WithRegion(parts[0]),
	)

	return &awsCfg, err
}

func importServiceMethods(moduleName string, connect any) ([]sdkmodule.Optfn, error) {
	connectv, connectt := reflect.ValueOf(connect), reflect.TypeOf(connect)
	if connectt.NumOut() != 1 {
		return nil, sdklogger.DPanicOrReturn("connect method must return only the client")
	}

	clientt := connectt.Out(0)
	if clientt.Kind() != reflect.Ptr || clientt.Elem().Kind() != reflect.Struct {
		return nil, sdklogger.DPanicOrReturn("client is not a pointer to a struct")
	}

	opts := make([]sdkmodule.Optfn, 0, clientt.NumMethod()+1)
	opts = append(opts, sdkmodule.WithDataFromConfig(func(config string) ([]byte, error) {
		_, err := getAWSConfig([]byte(config)) // validate.
		return []byte(config), err
	}))

	for mi := 0; mi < clientt.NumMethod(); mi++ {
		m := clientt.Method(mi)

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
				return nil, err
			}

			cfg, err := getAWSConfig(sdkmodule.FunctionDataFromContext(ctx))
			if err != nil {
				return nil, fmt.Errorf("token: %w", err)
			}

			if cfg == nil {
				return nil, fmt.Errorf("no config specified")
			}

			connectrets := connectv.Call([]reflect.Value{reflect.ValueOf(*cfg)})
			if len(connectrets) != 1 {
				return nil, fmt.Errorf("new client returned invalid values")
			}

			method := connectrets[0].MethodByName(m.Name) // must be original name.

			retvs := method.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				paramsValue,
			})

			if len(retvs) != 2 {
				return nil, fmt.Errorf("call returned %d values != expected 2", len(retvs))
			}

			outv, errv := retvs[0], retvs[1]

			if !errv.IsNil() {
				if err, ok := errv.Interface().(error); ok {
					return nil, err
				}

				return nil, fmt.Errorf("invalid error return")
			}

			out, err := sdkvalues.Wrap(outv.Interface())
			if err != nil {
				return nil, fmt.Errorf("return value conversion error: %w", err)
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
