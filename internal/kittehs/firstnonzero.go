package kittehs

// FirstNonZero returns first opt which is non-zero.
// If opts is empty, the zero value of T is returned.
func FirstNonZero[T comparable](opts ...T) (t T) {
	for _, opt := range opts {
		if opt != t {
			t = opt
			return
		}
	}
	return
}
