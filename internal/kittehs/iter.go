package kittehs

import "iter"

func IterToSlice[T any](seq iter.Seq[T]) []T {
	var out []T
	for v := range seq {
		out = append(out, v)
	}

	return out
}
