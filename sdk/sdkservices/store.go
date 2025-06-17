package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Store interface {
	Mutate(ctx context.Context, pid sdktypes.ProjectID, key, op string, operarnds ...sdktypes.Value) (sdktypes.Value, error)
	Get(ctx context.Context, pid sdktypes.ProjectID, keys []string) (map[string]sdktypes.Value, error)
	List(ctx context.Context, pid sdktypes.ProjectID) ([]string, error)
}
