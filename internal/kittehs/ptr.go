package kittehs

func ZeroIfNil[T any](t *T) (r T) {
	if t != nil {
		r = *t
	}

	return
}
