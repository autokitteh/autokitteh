package sdkprojectsclient

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	projectsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/projects/v1/projectsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client projectsv1connect.ProjectsServiceClient
}

func New(p sdkclient.Params) sdkservices.Projects {
	return &client{client: internal.New(projectsv1connect.NewProjectsServiceClient, p)}
}

func (c *client) Create(ctx context.Context, project sdktypes.Project) (sdktypes.ProjectID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&projectsv1.CreateRequest{
		Project: project.ToProto(),
	}))
	if err != nil {
		return sdktypes.InvalidProjectID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidProjectID, err
	}

	pid, err := sdktypes.StrictParseProjectID(resp.Msg.ProjectId)
	if err != nil {
		return sdktypes.InvalidProjectID, fmt.Errorf("invalid project id: %w", err)
	}

	return pid, nil
}

func (c *client) Delete(ctx context.Context, projectID sdktypes.ProjectID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(&projectsv1.DeleteRequest{ProjectId: projectID.String()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Update(ctx context.Context, project sdktypes.Project) error {
	resp, err := c.client.Update(ctx, connect.NewRequest(&projectsv1.UpdateRequest{
		Project: project.ToProto(),
	}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) GetByID(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.Project, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(
		&projectsv1.GetRequest{ProjectId: pid.String()},
	))
	if err != nil {
		return sdktypes.InvalidProject, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidProject, err
	}

	project, err := sdktypes.StrictProjectFromProto(resp.Msg.Project)
	if err != nil {
		// FIXME: ENG-626: why we check and override errInvalid for project only?
		var errInvalid sdkerrors.ErrInvalidArgument
		if err.Error() == "zero object" && errors.As(err, &errInvalid) {
			return sdktypes.InvalidProject, nil
		}
		return sdktypes.InvalidProject, err
	}

	return project, nil
}

func (c *client) GetByName(ctx context.Context, n sdktypes.Symbol) (sdktypes.Project, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(
		&projectsv1.GetRequest{
			Name: n.String(),
		},
	))
	if err != nil {
		return sdktypes.InvalidProject, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidProject, err
	}

	if resp.Msg.Project == nil {
		return sdktypes.InvalidProject, nil
	}

	project, err := sdktypes.ProjectFromProto(resp.Msg.Project)
	if err != nil {
		return sdktypes.InvalidProject, fmt.Errorf("invalid project: %w", err)
	}

	return project, nil
}

func (c *client) List(ctx context.Context) ([]sdktypes.Project, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&projectsv1.ListRequest{}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Projects, sdktypes.StrictProjectFromProto)
}

func (c *client) Build(ctx context.Context, pid sdktypes.ProjectID) (sdktypes.BuildID, error) {
	resp, err := c.client.Build(ctx, connect.NewRequest(
		&projectsv1.BuildRequest{ProjectId: pid.String()},
	))
	if err != nil {
		return sdktypes.InvalidBuildID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidBuildID, err
	}

	if resp.Msg.Error == nil {
		return sdktypes.StrictParseBuildID(resp.Msg.BuildId)
	}

	perr, err := sdktypes.ProgramErrorFromProto(resp.Msg.Error)
	if err != nil {
		return sdktypes.InvalidBuildID, err
	}

	return sdktypes.InvalidBuildID, perr.ToError()
}

func (c *client) SetResources(ctx context.Context, pid sdktypes.ProjectID, resources map[string][]byte) error {
	resp, err := c.client.SetResources(ctx, connect.NewRequest(
		&projectsv1.SetResourcesRequest{
			ProjectId: pid.String(),
			Resources: resources,
		},
	))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) DownloadResources(ctx context.Context, pid sdktypes.ProjectID) (map[string][]byte, error) {
	resp, err := c.client.DownloadResources(ctx, connect.NewRequest(
		&projectsv1.DownloadResourcesRequest{ProjectId: pid.String()},
	))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return resp.Msg.Resources, nil
}
