package sdkapplyclient

import (
	"context"

	"connectrpc.com/connect"

	applyv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1/applyv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type client struct {
	client applyv1connect.ApplyServiceClient
}

func (c *client) Apply(ctx context.Context, manifest, path string) ([]string, error) {
	resp, err := c.client.Apply(ctx, connect.NewRequest(&applyv1.ApplyRequest{Manifest: manifest, Path: path}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return resp.Msg.Logs, nil
}

func (c *client) Plan(ctx context.Context, manifest string) ([]string, error) {
	resp, err := c.client.Plan(ctx, connect.NewRequest(&applyv1.PlanRequest{Manifest: manifest}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return resp.Msg.Logs, nil
}

func New(p sdkclient.Params) sdkservices.Apply {
	return &client{client: internal.New(applyv1connect.NewApplyServiceClient, p)}
}
