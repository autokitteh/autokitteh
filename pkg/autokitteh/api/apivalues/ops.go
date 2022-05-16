package apivalues

import (
	"errors"
	"fmt"
)

var (
	ErrOutOfRange        = errors.New("out of range")
	ErrIncompatibleValue = errors.New("incompatible value")
)

func Length(v *Value) (int, error) {
	switch vv := v.Get().(type) {
	case NoneValue:
		return 0, nil
	case StringValue:
		return len(vv), nil
	case BytesValue:
		return len(vv), nil
	case ListValue:
		return len(vv), nil
	case SetValue:
		return len(vv), nil
	case DictValue:
		return len(vv), nil
	default:
		return 0, ErrIncompatibleValue
	}
}

func Index(v *Value, idx int) (*Value, error) {
	l, ok := v.Get().(ListValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}

	if idx < 0 {
		idx += len(l)
	}

	if idx < 0 || idx >= len(l) {
		return nil, ErrOutOfRange
	}

	return l[idx], nil
}

func Inc(v *Value, amount int64) (*Value, error) {
	vv, ok := v.Get().(IntegerValue)
	if !ok {
		return nil, errors.New("value must be an integer")
	}

	return Integer(int64(vv) + amount), nil
}

func Insert(v *Value, idx int, vv *Value) (*Value, error) {
	n, err := Length(v)
	if err != nil {
		return nil, err
	}

	g := v.Get()

	if s, ok := g.(SetValue); ok {
		if idx != 0 {
			return nil, errors.New("index must be zero for set insert")
		}

		return NewValue(SetValue(append(s, v)))
	} else if l, ok := g.(ListValue); ok {
		if idx < 0 {
			return NewValue(ListValue(append(l, vv)))
		} else {
			if idx > n {
				return nil, ErrOutOfRange
			}

			if idx == n {
				return NewValue(append(l, vv))
			}

			l = ListValue(append(l[:idx+1], l[idx:]...))
			l[idx] = vv
			return NewValue(l)
		}
	}

	return nil, ErrIncompatibleValue
}

func Take(v *Value, idx, count int) (*Value, []*Value, error) {
	l, ok := v.Get().(ListValue)
	if !ok {
		return nil, nil, ErrIncompatibleValue
	}

	if idx >= len(l) {
		return nil, nil, ErrOutOfRange
	}

	if count < 0 || idx+count > len(l) {
		count = len(l) - idx
	}

	taken := l[idx : idx+count]

	left := make([]*Value, len(l))
	copy(left, l) // see https://stackoverflow.com/posts/35276346/revisions
	left = append(left[:idx], left[idx+count:]...)

	return List(left...), taken, nil
}

func SetKey(v, kv, vv *Value) (*Value, error) {
	v = v.Clone()

	g := v.Get()

	d, ok := g.(DictValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}

	i := d.GetKey(kv)
	if i != nil {
		i.V = vv
		return Dict(d...), nil
	}

	return Dict(append(d, &DictItem{K: kv, V: vv})...), nil
}

func GetKey(v, k *Value) (*Value, error) {
	g := v.Get()

	d, ok := g.(DictValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}

	i := d.GetKey(k)
	if i == nil {
		return nil, nil
	}

	return i.V, nil
}

func Keys(v *Value) (*Value, error) {
	g := v.Get()

	d, ok := g.(DictValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}

	ks := make([]*Value, len(d))
	for i, kv := range d {
		ks[i] = kv.K
	}

	return List(ks...), nil
}

func Div(a, b *Value) (*Value, error) {
	switch va := a.Get().(type) {
	case IntegerValue:
		switch vb := b.Get().(type) {
		case IntegerValue:
			return Integer(int64(va) / int64(vb)), nil
		case FloatValue:
			return Float(float32(va) / float32(vb)), nil
		}
	case FloatValue:
		switch vb := b.Get().(type) {
		case FloatValue:
			return Float(float32(va) / float32(vb)), nil
		case IntegerValue:
			return Integer(int64(va) / int64(vb)), nil
		}
	}

	return nil, fmt.Errorf("operands must be numbers")
}

func Mul(a, b *Value) (*Value, error) {
	switch va := a.Get().(type) {
	case IntegerValue:
		switch vb := b.Get().(type) {
		case IntegerValue:
			return Integer(int64(va) * int64(vb)), nil
		case FloatValue:
			return Float(float32(va) * float32(vb)), nil
		}
	case FloatValue:
		switch vb := b.Get().(type) {
		case FloatValue:
			return Float(float32(va) * float32(vb)), nil
		case IntegerValue:
			return Integer(int64(va) * int64(vb)), nil
		}
	}

	return nil, fmt.Errorf("operands must be numbers")
}

func Add(a, b *Value) (*Value, error) {
	switch va := a.Get().(type) {
	case IntegerValue:
		switch vb := b.Get().(type) {
		case IntegerValue:
			return Integer(int64(va) + int64(vb)), nil
		case FloatValue:
			return Float(float32(va) + float32(vb)), nil
		}
	case FloatValue:
		switch vb := b.Get().(type) {
		case FloatValue:
			return Float(float32(va) + float32(vb)), nil
		case IntegerValue:
			return Integer(int64(va) + int64(vb)), nil
		}
	}

	return nil, fmt.Errorf("operands must be numbers")
}

func Foldr(f func(a, b *Value) (*Value, error), v0 *Value, vs []*Value) (*Value, error) {
	acc := v0

	for i, v := range vs {
		var err error
		if acc, err = f(acc, v); err != nil {
			return nil, fmt.Errorf("%d: %w", i, err)
		}
	}

	return acc, nil
}

// This mutates v.
func SetCallIssuer(v *Value, issuer string) error {
	c := v.pb.GetCall()
	if c == nil {
		return errors.New("not a call value")
	}

	c.Issuer = issuer
	return nil
}
