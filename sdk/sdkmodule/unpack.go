package sdkmodule

import (
	"fmt"
	"maps"
	"reflect"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// UnpackArgs unpacks the positional and keyword arguments into the supplied parameter
// variables. pairs is an alternating list of names and pointers to variables.
//
// If the parameter name ends with "?", it is optional.
// If the parameter name ends with "=", it must be supplied in kwargs.
// If the parameter name starts with "**", the destination will accept all kwargs as a dict.
// If the parameter name starts with "*", the destination will aceppt all args as a list.
//
// A nameless parameter can also be specified. That parameter must be a pointer to a struct.
// The function will use the member names of the struct as the parameter names. If the fields
// are tagged with `json:"..."`, the tag will be used as the parameter name. If the tag is
// "-", the field will be ignored. If the tag modifier is "omitempty", the field will be optional.
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
func UnpackArgs(args []sdktypes.Value, kwargs map[string]sdktypes.Value, dsts ...any) error {
	kwargs = maps.Clone(kwargs)

	var flattened []any

	for i := 0; i < len(dsts); i++ {
		if _, ok := dsts[i].(string); ok {
			i++
			continue
		}

		t := reflect.TypeOf(dsts[i])
		if t.Kind() != reflect.Ptr {
			return fmt.Errorf("dst %d must be name or a pointer to a struct", i)
		}

		tt := t.Elem()
		if tt.Kind() != reflect.Struct {
			return fmt.Errorf("dst %d must be a pointer to a struct", i)
		}

		for j := 0; j < tt.NumField(); j++ {
			ttf := tt.Field(j)

			name := ttf.Name
			optional := ttf.Type.Kind() == reflect.Ptr

			if j := ttf.Tag.Get("json"); j != "" {
				if j == "-" {
					continue
				}

				var rest string
				name, rest, _ = strings.Cut(j, ",")
				if rest == "omitempty" {
					optional = true
				}
			}

			if len(name) == 0 {
				continue
			}

			name += "="

			if optional {
				name += "?"
			}

			flattened = append(flattened, name, reflect.ValueOf(dsts[i]).Elem().Field(j).Addr().Interface())
		}

		dsts = append(dsts[:i], dsts[i+1:]...)
	}

	dsts = append(dsts, flattened...)

	for i := 0; i+1 < len(dsts); i += 2 {
		nameitf, dst := dsts[i], dsts[i+1]

		name, ok := nameitf.(string)
		if !ok {
			return fmt.Errorf("dst pair %d name must be a string", i/2)
		}

		if strings.HasPrefix(name, "**") {
			if err := sdktypes.DefaultValueWrapper.UnwrapInto(dst, sdktypes.NewDictValueFromStringMap(kwargs)); err != nil {
				return fmt.Errorf("dst %q: %w", name, err)
			}

			kwargs = nil
			continue
		} else if strings.HasPrefix(name, "*") {
			if err := sdktypes.DefaultValueWrapper.UnwrapInto(dst, kittehs.Must1(sdktypes.NewListValue(args))); err != nil {
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

		if err := sdktypes.DefaultValueWrapper.UnwrapInto(dst, v); err != nil {
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
