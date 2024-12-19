package policy

import (
	"context"
)

type DecideFunc func(ctx context.Context, path string, input any) (any, error)
