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

func (c *client) Get(ctx context.Context, id sdktypes.IntegrationID) (sdkservices.Integration, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(&integrationsv1.GetRequest{
		IntegrationId: id.String(),
	}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	if resp.Msg.Integration == nil {
		return nil, nil
	}

	desc, err := sdktypes.StrictIntegrationFromProto(resp.Msg.Integration)
	if err != nil {
		return nil, err
	}

	return &integration{desc: desc, client: c.client}, nil
}

func (c *client) List(ctx context.Context, nameSubstring string) ([]sdktypes.Integration, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(&integrationsv1.ListRequest{
		NameSubstring: nameSubstring,
	}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}
	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Integrations, sdktypes.StrictIntegrationFromProto)
}
