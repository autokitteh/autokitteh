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
	Get(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error)
	Create(ctx context.Context, deployment sdktypes.Deployment) (sdktypes.DeploymentID, error)
	List(ctx context.Context, filter ListDeploymentsFilter) ([]sdktypes.Deployment, error)
	Activate(ctx context.Context, deploymentID sdktypes.DeploymentID) error
	Test(ctx context.Context, deploymentID sdktypes.DeploymentID) error
	Drain(ctx context.Context, deploymentID sdktypes.DeploymentID) error
	Deactivate(ctx context.Context, deploymentID sdktypes.DeploymentID) error
}
