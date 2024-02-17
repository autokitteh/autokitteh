package sdkmappingclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	mappingsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/mappings/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/mappings/v1/mappingsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client mappingsv1connect.MappingsServiceClient
}

func New(p sdkclient.Params) sdkservices.Mappings {
	return &client{client: internal.New(mappingsv1connect.NewMappingsServiceClient, p)}
}

func (c *client) Create(ctx context.Context, mapping sdktypes.Mapping) (sdktypes.MappingID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&mappingsv1.CreateRequest{Mapping: mapping.ToProto()}))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	eventId, err := sdktypes.StrictParseMappingID(resp.Msg.MappingId)
	if err != nil {
		return nil, fmt.Errorf("invalid mapping id: %w", err)
	}

	return eventId, nil
}

func (c *client) Delete(ctx context.Context, mappingID sdktypes.MappingID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(
		&mappingsv1.DeleteRequest{MappingId: mappingID.String()},
	))
	if err != nil {
		return rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Get(ctx context.Context, mappingID sdktypes.MappingID) (sdktypes.Mapping, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(
		&mappingsv1.GetRequest{MappingId: mappingID.String()},
	))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return sdktypes.MappingFromProto(resp.Msg.Mapping)
}

func (c *client) List(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(
		&mappingsv1.ListRequest{EnvId: filter.EnvID.String(), ConnectionId: filter.ConnectionID.String()},
	))
	if err != nil {
		return nil, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Mappings, sdktypes.MappingFromProto)
}
