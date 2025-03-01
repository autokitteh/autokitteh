package sdkdeploymentsclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	deploymentsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/deployments/v1/deploymentsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client deploymentsv1connect.DeploymentsServiceClient
}

func New(p sdkclient.Params) sdkservices.Deployments {
	return &client{client: internal.New(deploymentsv1connect.NewDeploymentsServiceClient, p)}
}

// Create implements sdkservices.Deployments.
func (c *client) Create(ctx context.Context, deployment sdktypes.Deployment) (sdktypes.DeploymentID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&deploymentsv1.CreateRequest{Deployment: deployment.ToProto()}))
	if err != nil {
		return sdktypes.InvalidDeploymentID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidDeploymentID, err
	}

	id, err := sdktypes.Strict(sdktypes.ParseDeploymentID(resp.Msg.DeploymentId))
	if err != nil {
		return sdktypes.InvalidDeploymentID, fmt.Errorf("invalid deployment id: %w", err)
	}
	return id, nil
}

// Activate implements sdkservices.Deployments.
func (c *client) Activate(ctx context.Context, id sdktypes.DeploymentID) error {
	resp, err := c.client.Activate(ctx, connect.NewRequest(&deploymentsv1.ActivateRequest{DeploymentId: id.String()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

// Deactivate implements sdkservices.Deployments.
func (c *client) Deactivate(ctx context.Context, id sdktypes.DeploymentID) error {
	resp, err := c.client.Deactivate(ctx, connect.NewRequest(&deploymentsv1.DeactivateRequest{DeploymentId: id.String()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

// Test implements sdkservices.Deployments.
func (c *client) Test(ctx context.Context, id sdktypes.DeploymentID) error {
	resp, err := c.client.Test(ctx, connect.NewRequest(&deploymentsv1.TestRequest{DeploymentId: id.String()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

// Get implements sdkservices.Deployments.
func (c *client) Get(ctx context.Context, id sdktypes.DeploymentID) (sdktypes.Deployment, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&deploymentsv1.GetRequest{DeploymentId: id.String()}))
	if err != nil {
		return sdktypes.InvalidDeployment, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidDeployment, err
	}
	if resp.Msg.Deployment == nil {
		return sdktypes.InvalidDeployment, nil
	}

	deployment, err := sdktypes.StrictDeploymentFromProto(resp.Msg.Deployment)
	if err != nil {
		return sdktypes.InvalidDeployment, fmt.Errorf("invalid deployment: %w", err)
	}
	return deployment, nil
}

// List implements sdkservices.Deployments.
func (c *client) List(ctx context.Context, filter sdkservices.ListDeploymentsFilter) ([]sdktypes.Deployment, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&deploymentsv1.ListRequest{
		OrgId:               filter.OrgID.String(),
		ProjectId:           filter.ProjectID.String(),
		BuildId:             filter.BuildID.String(),
		State:               filter.State.ToProto(),
		Limit:               filter.Limit,
		IncludeSessionStats: filter.IncludeSessionStats,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	deployments, err := kittehs.TransformError(resp.Msg.Deployments, sdktypes.StrictDeploymentFromProto)
	if err != nil {
		return nil, err
	}

	return deployments, nil
}

func (c *client) Delete(ctx context.Context, deploymentID sdktypes.DeploymentID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(&deploymentsv1.DeleteRequest{DeploymentId: deploymentID.String()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}
