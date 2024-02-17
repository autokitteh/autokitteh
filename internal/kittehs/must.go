package kittehs

import (
	"fmt"
)

func Must0(err error) {
	if err != nil {
		Panic(err)
	}
}

func Must1[A any](a A, err error) A {
	if err != nil {
		Panic(err)
	}

	return a
}

func Must2[A, B any](a A, b B, err error) (A, B) {
	if err != nil {
		Panic(err)
	}

	return a, b
}

// Lazy Must1 with 1 input argument.
func Must11[A, X any](f func(A) (X, error)) func(A) X {
	return func(a A) X { return Must1(f(a)) }
}

// Lazy Must2 with 1 input argument.
func Must12[A, X, Y any](f func(A) (X, Y, error)) func(A) (X, Y) {
	return func(a A) (X, Y) { return Must2(f(a)) }
}

func MustEqual[T comparable](a, b T, msg ...any) {
	if a != b {
		Panic(fmt.Sprint(msg...))
	}
}

// If err is nil, return t. Else return alt.
func Should1[T any](alt T) func(t T, err error) T { return ShouldFunc1(func(error) T { return alt }) }

func ShouldFunc1[T any](alt func(error) T) func(t T, err error) T {
	return func(t T, err error) T {
		if err != nil {
			return alt(err)
		}

		return t
	}
}

func Should11[X, Y any](alt Y, f func(X) (Y, error)) func(X) Y {
	return ShouldFunc11(func(error) Y { return alt }, f)
}

// Same as Should1s, but alt is lazy.
func ShouldFunc11[X, Y any](alt func(error) Y, f func(X) (Y, error)) func(X) Y {
	return func(x X) Y { return ShouldFunc1(alt)(f(x)) }
}
