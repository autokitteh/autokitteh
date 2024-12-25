package sdktriggerclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	triggersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1/triggersv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client triggersv1connect.TriggersServiceClient
}

func New(p sdkclient.Params) sdkservices.Triggers {
	return &client{client: internal.New(triggersv1connect.NewTriggersServiceClient, p)}
}

func (c *client) Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error) {
	resp, err := c.client.Create(ctx, connect.NewRequest(&triggersv1.CreateRequest{Trigger: trigger.ToProto()}))
	if err != nil {
		return sdktypes.InvalidTriggerID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	triggerID, err := sdktypes.StrictParseTriggerID(resp.Msg.TriggerId)
	if err != nil {
		return sdktypes.InvalidTriggerID, fmt.Errorf("invalid trigger id: %w", err)
	}

	return triggerID, nil
}

func (c *client) Update(ctx context.Context, trigger sdktypes.Trigger) error {
	resp, err := c.client.Update(ctx, connect.NewRequest(&triggersv1.UpdateRequest{Trigger: trigger.ToProto()}))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Delete(ctx context.Context, triggerID sdktypes.TriggerID) error {
	resp, err := c.client.Delete(ctx, connect.NewRequest(
		&triggersv1.DeleteRequest{TriggerId: triggerID.String()},
	))
	if err != nil {
		return rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return err
	}

	return nil
}

func (c *client) Get(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(
		&triggersv1.GetRequest{TriggerId: triggerID.String()},
	))
	if err != nil {
		return sdktypes.InvalidTrigger, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidTrigger, err
	}

	return sdktypes.TriggerFromProto(resp.Msg.Trigger)
}

func (c *client) List(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(
		&triggersv1.ListRequest{
			ConnectionId: filter.ConnectionID.String(),
			ProjectId:    filter.ProjectID.String(),
			SourceType:   filter.SourceType.ToProto(),
		},
	))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Triggers, sdktypes.TriggerFromProto)
}
