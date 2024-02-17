package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Store interface {
	Get(ctx context.Context, envID sdktypes.EnvID, projectID sdktypes.ProjectID, keys []string) (map[string]sdktypes.Value, error)
	List(ctx context.Context, envID sdktypes.EnvID, projectID sdktypes.ProjectID) ([]string, error)
}
