package kittehs

// Return first opt which is non-zero.
func Choose[T comparable](opts ...T) (t T) {
	for _, opt := range opts {
		if opt != t {
			t = opt
			return
		}
	}
	return
}
