package kittehs

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
