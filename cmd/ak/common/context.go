package common

import (
	"context"
	"fmt"
	"os"
	"time"
)

var timeout = 1 * time.Minute

func init() {
	if v, ok := os.LookupEnv("AK_TIMEOUT"); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			panic(fmt.Errorf("invalid AK_TIMEOUT: %q, %w", v, err))
		}

		timeout = d
	}
}

func LimitedContext() (context.Context, context.CancelFunc) {
	return WithLimitedContext(context.Background())
}

// TODO(ENG-320): Configuration to disable timeout for debugging.
func WithLimitedContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, timeout)
}
