package sdkbuildsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	buildsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/builds/v1/buildsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client buildsv1connect.BuildsServiceClient
}

// Download implements sdkservices.Builds.
func (c *client) Download(ctx context.Context, buildID sdktypes.BuildID) (io.ReadCloser, error) {
	resp, err := c.client.Download(ctx, connect.NewRequest(&buildsv1.DownloadRequest{BuildId: buildID.String()}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	reader := io.NopCloser(bytes.NewReader(resp.Msg.Data))
	return reader, nil
}

// Get implements sdkservices.Builds.
func (c *client) Get(ctx context.Context, buildID sdktypes.BuildID) (sdktypes.Build, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&buildsv1.GetRequest{BuildId: buildID.String()}))
	if err != nil {
		return sdktypes.InvalidBuild, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidBuild, err
	}
	if resp.Msg.Build == nil {
		return sdktypes.InvalidBuild, nil
	}

	build, err := sdktypes.StrictBuildFromProto(resp.Msg.Build)
	if err != nil {
		return sdktypes.InvalidBuild, err
	}
	return build, nil
}

// List implements sdkservices.Builds.
func (c *client) List(ctx context.Context, filter sdkservices.ListBuildsFilter) ([]sdktypes.Build, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&buildsv1.ListRequest{
		Limit: filter.Limit,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	builds, err := kittehs.TransformError(resp.Msg.Builds, sdktypes.StrictBuildFromProto)
	if err != nil {
		return nil, fmt.Errorf("invalid build: %w", err)
	}
	return builds, nil
}

// Remove implements sdkservices.Builds.
func (c *client) Delete(ctx context.Context, buildID sdktypes.BuildID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(&buildsv1.DeleteRequest{BuildId: buildID.String()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

// Save implements sdkservices.Builds.
func (c *client) Save(ctx context.Context, build sdktypes.Build, data []byte) (sdktypes.BuildID, error) {
	resp, err := c.client.Save(ctx, connect.NewRequest(&buildsv1.SaveRequest{Build: build.ToProto(), Data: data}))
	if err != nil {
		return sdktypes.InvalidBuildID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidBuildID, err
	}

	buildID, err := sdktypes.StrictParseBuildID(resp.Msg.BuildId)
	if err != nil {
		return sdktypes.InvalidBuildID, fmt.Errorf("invalid build: %w", err)
	}
	return buildID, nil
}

func (c *client) Describe(ctx context.Context, buildID sdktypes.BuildID) (*sdkbuildfile.BuildFile, error) {
	resp, err := c.client.Describe(ctx, connect.NewRequest(&buildsv1.DescribeRequest{BuildId: buildID.String()}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	var bf sdkbuildfile.BuildFile

	if err := json.Unmarshal([]byte(resp.Msg.DescriptionJson), &bf); err != nil {
		return nil, fmt.Errorf("invalid description: %w", err)
	}

	return &bf, nil
}

func New(p sdkclient.Params) sdkservices.Builds {
	return &client{client: internal.New(buildsv1connect.NewBuildsServiceClient, p)}
}
