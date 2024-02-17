package kittehs

import (
	"sync"
)

func Lazy1[T, P any](f func(P) T, p P) func() T {
	var (
		t    T
		once sync.Once
	)

	return func() T {
		once.Do(func() { t = f(p) })

		return t
	}
}
