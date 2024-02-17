package values

import (
	"errors"
	"fmt"
	"time"

	starlarktime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func Unwrap(v starlark.Value) (any, error) {
	if v == nil {
		return nil, nil
	}

	switch v := v.(type) {
	case starlark.Bool:
		return bool(v), nil
	case starlark.Bytes:
		return string(v), nil
	case starlark.String:
		return string(v), nil
	case starlark.Float:
		return float64(v), nil
	case starlark.Int:
		if i64, ok := v.Int64(); ok {
			return i64, nil
		}
		return nil, errors.New("cannot convert starlark int")
	case starlark.NoneType:
		return nil, nil
	case starlarktime.Time:
		return time.Time(v), nil
	case starlarktime.Duration:
		return time.Duration(v), nil
	case *Symbol:
		return v.String(), nil
	case *starlark.Set, *starlark.List, starlark.Tuple:
		seq := v.(starlark.Sequence)
		iter := seq.Iterate()
		defer iter.Done()

		l := make([]any, seq.Len())

		var dst starlark.Value
		for i := 0; iter.Next(&dst); i++ {
			var err error
			if l[i], err = Unwrap(dst); err != nil {
				return nil, fmt.Errorf("seq[%d]: %w", i, err)
			}
		}
		return l, nil
	case *starlarkstruct.Struct:
		strd := make(starlark.StringDict)
		v.ToStringDict(strd)
		dd := starlark.NewDict(len(strd))
		for k, v := range strd {
			if err := dd.SetKey(starlark.String(k), v); err != nil {
				return nil, fmt.Errorf("set key %q: %w", k, err)
			}
		}
		return dd, nil
	case *starlarkstruct.Module:
		dd := starlark.NewDict(len(v.Members))
		for k, v := range v.Members {
			if err := dd.SetKey(starlark.String(k), v); err != nil {
				return nil, fmt.Errorf("set key %q: %w", k, err)
			}
		}
		return dd, nil
	case *starlark.Dict:
		m := make(map[any]any, v.Len())
		for _, kv := range v.Items() {
			k, err := Unwrap(kv[0])
			if err != nil {
				return nil, fmt.Errorf("key: %w", err)
			}

			v, err := Unwrap(kv[0])
			if err != nil {
				return nil, fmt.Errorf("value: %w", err)
			}

			m[k] = v
		}
		return m, nil
	}

	return nil, fmt.Errorf("cannot convert starlark %q", v.Type())
}
