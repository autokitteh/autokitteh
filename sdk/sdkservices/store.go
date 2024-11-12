package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Store interface {
	Get(ctx context.Context, pid sdktypes.ProjectID, keys []string) (map[string]sdktypes.Value, error)
	List(ctx context.Context, pid sdktypes.ProjectID) ([]string, error)
}
