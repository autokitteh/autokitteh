package sdkruntimesclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1/runtimesv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client runtimesv1connect.RuntimesServiceClient
}

func New(p sdkclient.Params) sdkservices.Runtimes {
	return &client{client: internal.New(runtimesv1connect.NewRuntimesServiceClient, p)}
}

func (c *client) List(ctx context.Context) ([]sdktypes.Runtime, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&runtimesv1.ListRequest{}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Runtimes, sdktypes.StrictRuntimeFromProto)
}

func (c *client) New(ctx context.Context, name sdktypes.Name) (sdkservices.Runtime, error) {
	resp, err := c.client.Describe(ctx, connect.NewRequest(&runtimesv1.DescribeRequest{Name: name.String()}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	if resp.Msg.Runtime == nil {
		return nil, nil
	}

	desc, err := sdktypes.StrictRuntimeFromProto(resp.Msg.Runtime)
	if err != nil {
		return nil, fmt.Errorf("invalid runtime: %w", err)
	}

	return &runtime{desc: desc}, nil
}
