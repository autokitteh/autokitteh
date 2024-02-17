package kittehs

import (
	"cmp"
	"slices"

	"golang.org/x/exp/maps"
)

func ValidateMap[K comparable, V any](m map[K]V, f func(K, V) error) error {
	for k, v := range m {
		if err := f(k, v); err != nil {
			return err
		}
	}
	return nil
}

// MapValuesSortedByKeys returns a slice of the map's values in stable order.
func MapValuesSortedByKeys[A cmp.Ordered, B any](m map[A]B) []B {
	ks := maps.Keys(m)
	slices.Sort(ks)
	vs := make([]B, len(m))
	for i, k := range ks {
		vs[i] = m[k]
	}
	return vs
}

func JoinMaps[K comparable, V any](ms ...map[K]V) (map[K]V, map[K]bool) {
	m := make(map[K]V)
	overrides := make(map[K]bool)
	for _, m1 := range ms {
		for k, v := range m1 {
			if _, ok := m[k]; ok {
				overrides[k] = true
			}
			m[k] = v
		}
	}
	return m, overrides
}

func MapFilter[K comparable, V any](m map[K]V, f func(K, V) bool) map[K]V {
	return NewMapFilter(f)(m)
}

func NewMapFilter[K comparable, V any](f func(K, V) bool) func(map[K]V) map[K]V {
	return func(m map[K]V) map[K]V {
		r := make(map[K]V, len(m)) // not sure len(m) is right here. f may be very selective.
		for k, v := range m {
			if f(k, v) {
				r[k] = v
			}
		}
		return r
	}
}

func NewMapKeysFilter[K comparable, V any](f func(K) bool) func(map[K]V) map[K]V {
	return NewMapFilter(func(k K, _ V) bool { return f(k) })
}

func FilterMapKeys[K comparable, V any](m map[K]V, f func(K) bool) map[K]V {
	return NewMapKeysFilter[K, V](func(k K) bool { return f(k) })(m)
}

func NewMapValuesFilter[K comparable, V any](f func(V) bool) func(map[K]V) map[K]V {
	return NewMapFilter(func(_ K, v V) bool { return f(v) })
}

func FilterMapValues[K comparable, V any](m map[K]V, f func(V) bool) map[K]V {
	return NewMapFilter[K, V](func(_ K, v V) bool { return f(v) })(m)
}
