package flexcall

import (
	"errors"
	"fmt"
	"reflect"
)

var ErrUnmatched = errors.New("could not find a match for argument")

// Call given function f with arguments args. Arguments will be supplied
// to f based on their types. If any given argument does not match -
// no error is returned. If any argument expected by f is not matched,
// an error wrapping ErrUnmatched is returned and f is not called.
// See tests for examples.
func Call(f interface{}, args ...interface{}) ([]interface{}, error) {
	return call(false, f, args...)
}

// CallOptional behaves the same way as Call, but if any value is unfulfilled,
// its zero value will be set instead of erroring out.
func CallOptional(f interface{}, args ...interface{}) ([]interface{}, error) {
	return call(true, f, args...)
}

func call(allowNils bool, f interface{}, args ...interface{}) ([]interface{}, error) {
	ft, fv := reflect.TypeOf(f), reflect.ValueOf(f)

	if ft.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a func")
	}

	avs := make([]reflect.Value, 0, len(args))
	for _, a := range args {
		avs = append(avs, reflect.ValueOf(a))
	}

	ins := make([]reflect.Value, ft.NumIn())

outer:
	for i := 0; i < ft.NumIn(); i++ {
		at := ft.In(i)

		for _, av := range avs {
			if at.Kind() == reflect.Interface {
				// An interface might fulfill multiple types, and that fails
				// a simple comparison with underlying types (the av.Type() == at branch).
				// Better check if the required arg can receive this value.
				if av.Type().Implements(at) {
					ins[i] = av
					continue outer
				}
			} else if av.Type() == at {
				ins[i] = av
				continue outer
			}
		}

		if allowNils {
			ins[i] = reflect.Zero(at)
			continue
		}

		return nil, fmt.Errorf(`%w: arg %d of type "%s"`, ErrUnmatched, i, at.String())
	}

	outs := fv.Call(ins)

	if len(outs) == 0 {
		return nil, nil
	}

	iouts := make([]interface{}, len(outs))
	for i, out := range outs {
		iouts[i] = out.Interface()
	}

	return iouts, nil
}

// ExtractError splits the input arguments into non-error and
// error parts. An error is considered only if it is the last
// item in the given arguments.
// f is the function which its return values are supplied in outs.
// This is necessary as an nil error does not carry a type in outs.
func ExtractError(f interface{}, outs []interface{}) ([]interface{}, error) {
	ft := reflect.TypeOf(f)

	if ft.NumOut() != len(outs) {
		panic("given values len not correspond with function return values len")
	}

	if ft.NumOut() == 0 {
		return nil, nil
	}

	lt := ft.Out(ft.NumOut() - 1)

	if lt.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		l := outs[len(outs)-1]

		if l == nil {
			return outs[:len(outs)-1], nil
		}

		return outs[:len(outs)-1], l.(error)
	}

	return outs, nil
}

func CallAndExtractError(f interface{}, args ...interface{}) ([]interface{}, error) {
	outs, err := Call(f, args...)
	if err != nil {
		return nil, err
	}

	return ExtractError(f, outs)
}

func CallOptionalAndExtractError(f interface{}, args ...interface{}) ([]interface{}, error) {
	outs, err := CallOptional(f, args...)
	if err != nil {
		return nil, err
	}

	return ExtractError(f, outs)
}
