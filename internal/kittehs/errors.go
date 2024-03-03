package kittehs

import "fmt"

func ErrorWithPrefix[T any](prefix T, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%v: %w", prefix, err)
}
