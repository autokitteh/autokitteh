package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListDeploymentsFilter struct {
	EnvID               sdktypes.EnvID
	BuildID             sdktypes.BuildID
	State               sdktypes.DeploymentState
	Limit               uint32
	IncludeSessionStats bool
}

type Deployments interface {
	Create(ctx context.Context, deployment sdktypes.Deployment) (sdktypes.DeploymentID, error)
	Activate(ctx context.Context, deploymentID sdktypes.DeploymentID) error
	Deactivate(ctx context.Context, deploymentID sdktypes.DeploymentID) error
	Get(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error)
	Test(ctx context.Context, deploymentID sdktypes.DeploymentID) error
	List(ctx context.Context, filter ListDeploymentsFilter) ([]sdktypes.Deployment, error)
	Delete(ctx context.Context, id sdktypes.DeploymentID) error
}
