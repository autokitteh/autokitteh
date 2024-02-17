package kittehs

func IsNil[T any](x *T) bool { return x == nil }

func IsNotNil[T any](x *T) bool { return x != nil }
