package kittehs

import "fmt"

// ErrorWithPrefix wraps the given error by prepending the prefix
// to it. If the error is nil, it also returns nil. The prefix
// comes before the error and not after, because it contextualizes
// the error. This still allows for additional error wrapping.
// Example: "config error: failed to read file: file not found".
func ErrorWithPrefix(prefix string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", prefix, err)
}

// ErrorWithValue wraps the given error by appending the value to
// it. If the error is nil, it also returns nil. The error comes
// before the value and not after, because it describes what's
// wrong with the value. This allows additional error wrapping.
// Example: "loop error: counter must be positive: -1".
func ErrorWithValue[T any](value T, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", err, value)
}
