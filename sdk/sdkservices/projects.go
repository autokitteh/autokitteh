package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Projects interface {
	Create(ctx context.Context, project sdktypes.Project) (sdktypes.ProjectID, error)
	GetByID(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.Project, error)
	GetByName(ctx context.Context, name sdktypes.Name) (sdktypes.Project, error)
	List(ctx context.Context) ([]sdktypes.Project, error)
	Build(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.BuildID, error)
	SetResources(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) error
	DownloadResources(ctx context.Context, projectID sdktypes.ProjectID) (map[string][]byte, error)
}
