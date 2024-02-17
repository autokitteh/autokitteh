package sdkmodule

import (
	"fmt"
	"maps"
	"strings"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/sdk/sdkvalues"
)

// UnpackArgs unpacks the positional and keyword arguments into the supplied parameter
// variables. pairs is an alternating list of names and pointers to variables.
//
// If the parameter name ends with "?", it is optional.
// If the parameter name ends with "=", it must be supplied in kwargs.
// If the parameter name starts with "**", the destination will accept all kwargs as a dict.
// If the parameter name starts with "*", the destination will aceppt all args as a list.
//
// Example:
//
//		func SomeFunc(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
//			var (
//			  x int
//			  y string
//			  args []int
//			)
//
//			if err := UnpackArgs(args, kwargs, "x", &x, "y=", &y, "*args", &args); err != nil {
//		  		return err
//			}
//
//	        ...
//		}
//
// (this function is heavily inspired by https://pkg.go.dev/go.starlark.net/starlark#UnpackArgs,
// it essentially does the same thing, but on autokitteh level and not starlark)
//
// TODO: varargs.
func UnpackArgs(args []sdktypes.Value, kwargs map[string]sdktypes.Value, dsts ...interface{}) error {
	kwargs = maps.Clone(kwargs)

	if len(dsts)%2 != 0 {
		return fmt.Errorf("must have even number of dsts")
	}

	for i := 0; i < len(dsts); i += 2 {
		nameitf, dst := dsts[i], dsts[i+1]

		name, ok := nameitf.(string)
		if !ok {
			return fmt.Errorf("dst pair %d name must be a string", i/2)
		}

		if strings.HasPrefix(name, "**") {
			if err := sdkvalues.DefaultValueWrapper.UnwrapInto(dst, sdktypes.NewDictValueFromStringMap(kwargs)); err != nil {
				return fmt.Errorf("dst %q: %w", name, err)
			}

			kwargs = nil
			continue
		} else if strings.HasPrefix(name, "*") {
			if err := sdkvalues.DefaultValueWrapper.UnwrapInto(dst, sdktypes.NewListValue(args)); err != nil {
				return fmt.Errorf("dst %q: %w", name, err)
			}

			args = nil
			continue
		}

		optional := strings.ContainsRune(name, '?')
		mustkw := strings.ContainsRune(name, '=')
		name = strings.TrimRight(name, "?=")

		v, found := kwargs[name]
		if found {
			delete(kwargs, name)
		} else {
			if len(args) > 0 && !mustkw {
				v, args = args[0], args[1:]
			} else {
				if !optional {
					return fmt.Errorf("required parameter %q not specified", name)
				}

				continue
			}
		}

		if err := sdkvalues.DefaultValueWrapper.UnwrapInto(dst, v); err != nil {
			return fmt.Errorf("dst %q: %w", name, err)
		}
	}

	if len(args) > 0 {
		return fmt.Errorf("not all positional arguments consumed: %v", args)
	}

	if len(kwargs) > 0 {
		return fmt.Errorf("not all keyword arguments consumed: %v", kwargs)
	}

	return nil
}
