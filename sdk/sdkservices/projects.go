package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Projects interface {
	Create(ctx context.Context, project sdktypes.Project) (sdktypes.ProjectID, error)
	Delete(ctx context.Context, projectID sdktypes.ProjectID) error
	Update(ctx context.Context, project sdktypes.Project) error
	GetByID(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.Project, error)
	GetByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.Project, error)
	List(ctx context.Context) ([]sdktypes.Project, error)
	Build(ctx context.Context, projectID sdktypes.ProjectID) (sdktypes.BuildID, error)
	SetResources(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte) error
	DownloadResources(ctx context.Context, projectID sdktypes.ProjectID) (map[string][]byte, error)
	Export(ctx context.Context, projectID sdktypes.ProjectID) ([]byte, error)
	Lint(ctx context.Context, projectID sdktypes.ProjectID, resources map[string][]byte, manifestPath string) ([]*sdktypes.CheckViolation, error)
}
