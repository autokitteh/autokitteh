package kittehs

import (
	"fmt"
)

func Transform[A, B any](as []A, f func(A) B) []B {
	if as == nil {
		return nil
	}

	bs := make([]B, len(as))
	for i, a := range as {
		bs[i] = f(a)
	}
	return bs
}

func TransformWithIndex[A, B any](as []A, f func(int, A) B) []B {
	if as == nil {
		return nil
	}

	bs := make([]B, len(as))
	for i, a := range as {
		bs[i] = f(i, a)
	}
	return bs
}

func TransformError[A, B any](as []A, f func(A) (B, error)) ([]B, error) {
	if as == nil {
		return nil, nil
	}

	bs := make([]B, len(as))
	for i, a := range as {
		var err error
		if bs[i], err = f(a); err != nil {
			return nil, fmt.Errorf("%w: %v", err, a)
		}
	}

	return bs, nil
}

func TransformMapToList[K comparable, V, T any](m map[K]V, f func(K, V) T) []T {
	l := make([]T, 0, len(m))
	for k, v := range m {
		l = append(l, f(k, v))
	}
	return l
}

func TransformMapToListError[K comparable, V, T any](m map[K]V, f func(K, V) (T, error)) ([]T, error) {
	l := make([]T, 0, len(m))
	for k, v := range m {
		i, err := f(k, v)
		if err != nil {
			return nil, fmt.Errorf("%w: key %v, value %v", err, k, v)
		}
		l = append(l, i)
	}
	return l, nil
}

func TransformMap[A0, A1 comparable, B0, B1 any](m map[A0]B0, f func(A0, B0) (A1, B1)) map[A1]B1 {
	m1 := make(map[A1]B1, len(m))
	for k, v := range m {
		k1, v1 := f(k, v)
		m1[k1] = v1
	}
	return m1
}

func TransformMapError[A0, A1 comparable, B0, B1 any](m map[A0]B0, f func(A0, B0) (A1, B1, error)) (map[A1]B1, error) {
	m1 := make(map[A1]B1, len(m))
	for k, v := range m {
		k1, v1, err := f(k, v)
		if err != nil {
			return nil, fmt.Errorf("%w: key %v, value %v", err, k, v)
		}
		m1[k1] = v1
	}
	return m1, nil
}

func TransformMapValues[A comparable, B0, B1 any](m map[A]B0, f func(B0) B1) map[A]B1 {
	return TransformMap(m, func(a A, b B0) (A, B1) { return a, f(b) })
}

func TransformMapValuesError[A comparable, B0, B1 any](m map[A]B0, f func(B0) (B1, error)) (map[A]B1, error) {
	return TransformMapError(m, func(a A, b B0) (A, B1, error) {
		b1, err := f(b)
		return a, b1, err
	})
}

func TransformToStrings[T fmt.Stringer](ts []T) []string {
	return Transform(ts, func(t T) string { return ToString(t) })
}

func TransformUnptr[T any](in []*T) (out []T) {
	if in == nil {
		return nil
	}

	out = make([]T, len(in))
	for i, t := range in {
		out[i] = *t
	}
	return
}

func TransformPtr[T any](in []T) (out []*T) {
	if in == nil {
		return nil
	}

	out = make([]*T, len(in))
	for i, t := range in {
		out[i] = func(t T) *T { return &t }(t)
	}
	return
}
