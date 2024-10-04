package kittehs

func Filter[T any](ts []T, f func(T) bool) []T {
	return NewFilter(f)(ts)
}

func FilterNils[T any](ts []*T) []*T { return Filter(ts, IsNotNil) }

func FilterZeroes[T comparable](ts []T) []T { return Filter(ts, IsNotZero) }

func NewFilter[T any](f func(T) bool) func([]T) []T {
	return func(xs []T) []T {
		if xs == nil {
			return nil
		}

		ys := []T{}
		for _, x := range xs {
			if f(x) {
				ys = append(ys, x)
			}
		}
		return ys
	}
}

func ContainedIn[T comparable](xs ...T) func(T) bool {
	m := make(map[T]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}

	return func(t T) bool { return m[t] }
}

// This does not guard against duplicate keys.
func ListToMap[T any, K comparable, V any](ts []T, f func(T) (K, V)) map[K]V {
	m := make(map[K]V, len(ts))
	for _, t := range ts {
		k, v := f(t)
		m[k] = v
	}
	return m
}

func ListToMapError[T any, K comparable, V any](ts []T, f func(T) (K, V, error)) (map[K]V, error) {
	m := make(map[K]V, len(ts))
	for _, t := range ts {
		k, v, err := f(t)
		if err != nil {
			return nil, err
		}
		m[k] = v
	}
	return m, nil
}

// Returns first index with error.
func ValidateList[T any](vs []T, f func(int, T) error) (int, error) {
	for i, v := range vs {
		if err := f(i, v); err != nil {
			return i, err
		}
	}

	return -1, nil
}

func FindFirst[T any](vs []T, f func(T) bool) (int, T) {
	var t T

	for i, v := range vs {
		if f(v) {
			return i, v
		}
	}

	return -1, t
}

func All(xs ...bool) bool {
	for _, x := range xs {
		if !x {
			return false
		}
	}

	return true
}
