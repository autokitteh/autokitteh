package sdkintegrationsclient

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	integrationsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/integrations/v1/integrationsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client integrationsv1connect.IntegrationsServiceClient
}

func New(p sdkclient.Params) sdkservices.Integrations {
	return &client{client: internal.New(integrationsv1connect.NewIntegrationsServiceClient, p)}
}

func (c *client) get(ctx context.Context, id sdktypes.IntegrationID, name sdktypes.Symbol) (sdktypes.Integration, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&integrationsv1.GetRequest{
		IntegrationId: id.String(),
		Name:          name.String(),
	}))
	if err != nil {
		return sdktypes.InvalidIntegration, rpcerrors.ToSDKError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidIntegration, err
	}

	if resp.Msg.Integration == nil {
		return sdktypes.InvalidIntegration, nil
	}

	desc, err := sdktypes.StrictIntegrationFromProto(resp.Msg.Integration)
	if err != nil {
		return sdktypes.InvalidIntegration, err
	}

	return desc, nil
}

func (c *client) Attach(ctx context.Context, id sdktypes.IntegrationID) (sdkservices.Integration, error) {
	desc, err := c.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &integration{desc: desc, client: c.client}, nil
}

func (c *client) GetByID(ctx context.Context, id sdktypes.IntegrationID) (sdktypes.Integration, error) {
	return c.get(ctx, id, sdktypes.InvalidSymbol)
}

func (c *client) GetByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.Integration, error) {
	return c.get(ctx, sdktypes.InvalidIntegrationID, name)
}

func (c *client) List(ctx context.Context, nameSubstring string) ([]sdktypes.Integration, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&integrationsv1.ListRequest{
		NameSubstring: nameSubstring,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Integrations, sdktypes.StrictIntegrationFromProto)
}
