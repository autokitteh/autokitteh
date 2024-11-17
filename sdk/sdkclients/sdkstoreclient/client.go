package sdkstoreclient

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	storev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1/storev1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client storev1connect.StoreServiceClient
}

func New(p sdkclient.Params) sdkservices.Store {
	return &client{client: internal.New(storev1connect.NewStoreServiceClient, p)}
}

func (c *client) List(ctx context.Context, pid sdktypes.ProjectID) ([]string, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&storev1.ListRequest{
		ProjectId: pid.String(),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return resp.Msg.Keys, nil
}

func (c *client) Get(ctx context.Context, pid sdktypes.ProjectID, keys []string) (map[string]sdktypes.Value, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&storev1.GetRequest{
		ProjectId: pid.String(),
		Keys:      keys,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformMapValuesError(resp.Msg.Values, sdktypes.StrictValueFromProto)
}
