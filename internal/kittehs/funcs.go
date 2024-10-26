package kittehs

import (
	"sync"
)

// LazyCache wraps a function and a single input. The first call to the wrapper
// calls the wrapped function. Subsequent calls return the first result.
func LazyCache[T, P any](f func(P) T, p P) func() T {
	var (
		t    T
		once sync.Once
	)

	return func() T {
		once.Do(func() { t = f(p) })

		return t
	}
}
