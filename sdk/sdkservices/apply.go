package sdkservices

import (
	"context"
)

type Apply interface {
	Apply(ctx context.Context, manifest string, path string) ([]string, error)
	Plan(ctx context.Context, manifest string) ([]string, error)
}
