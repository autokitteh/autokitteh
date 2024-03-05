package kittehs

func IsNil[T any](x *T) bool    { return x == nil }
func IsNotNil[T any](x *T) bool { return x != nil }

func IsZero[T comparable](t T) bool    { var zero T; return t == zero }
func IsNotZero[T comparable](t T) bool { var zero T; return t != zero }
