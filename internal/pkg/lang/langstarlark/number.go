package langstarlark

import (
	"fmt"

	"go.starlark.net/starlark"
)

type number struct {
	i *starlark.Int
	f *starlark.Float
}

func (n *number) Unpack(v starlark.Value) error {
	switch vv := v.(type) {
	case starlark.Int:
		n.i = &vv
	case starlark.Float:
		n.f = &vv
	default:
		return fmt.Errorf("not a number: %v", v.Type())
	}

	return nil
}

func (n *number) AsFloat() float64 {
	if n.f != nil {
		return float64(*n.f)
	}

	i64, ok := n.i.Int64()
	if !ok {
		panic("int64 conversion error")
	}

	return float64(i64)
}
