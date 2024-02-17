package common

import (
	"context"
	"time"
)

const (
	timeout = 1 * time.Minute
)

// TODO(ENG-320): Configuration to disable timeout for debugging.
func LimitedContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
