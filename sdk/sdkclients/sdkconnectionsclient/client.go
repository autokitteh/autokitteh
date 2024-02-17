package sdkconnectionsclient

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	connectionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/connections/v1/connectionsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client connectionsv1connect.ConnectionsServiceClient
}

func New(p sdkclient.Params) sdkservices.Connections {
	return &client{client: internal.New(connectionsv1connect.NewConnectionsServiceClient, p)}
}

func (c *client) Update(ctx context.Context, conn sdktypes.Connection) error {
	resp, err := c.client.Update(ctx, connect.NewRequest(&connectionsv1.UpdateRequest{
		Connection: conn.ToProto(),
	}))
	if err != nil {
		return rpcerrors.TranslateError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Create(ctx context.Context, conn sdktypes.Connection) (sdktypes.ConnectionID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&connectionsv1.CreateRequest{
		Connection: conn.ToProto(),
	}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return sdktypes.StrictParseConnectionID(resp.Msg.ConnectionId)
}

func (c *client) Delete(ctx context.Context, id sdktypes.ConnectionID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(&connectionsv1.DeleteRequest{
		ConnectionId: id.String(),
	}))
	if err != nil {
		return rpcerrors.TranslateError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Get(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&connectionsv1.GetRequest{
		ConnectionId: id.String(),
	}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	if resp.Msg.Connection == nil {
		return nil, nil
	}

	return sdktypes.StrictConnectionFromProto(resp.Msg.Connection)
}

func (c *client) List(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&connectionsv1.ListRequest{
		IntegrationId:    filter.IntegrationID.String(),
		IntegrationToken: filter.IntegrationToken,
		ProjectId:        filter.ProjectID.String(),
	}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Connections, sdktypes.StrictConnectionFromProto)
}
