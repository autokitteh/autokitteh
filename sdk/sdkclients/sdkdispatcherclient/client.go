package sdkdispatcherclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	dispatcherv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1/dispatcherv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client dispatcherv1connect.DispatcherServiceClient
}

func New(p sdkclient.Params) sdkservices.Dispatcher {
	return &client{client: internal.New(dispatcherv1connect.NewDispatcherServiceClient, p)}
}

func (c *client) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	resp, err := c.client.Redispatch(
		ctx,
		connect.NewRequest(
			&dispatcherv1.RedispatchRequest{
				EventId:      eventID.String(),
				DeploymentId: opts.DeploymentID.String(),
				EnvId:        opts.Env,
			},
		),
	)
	if err != nil {
		return sdktypes.InvalidEventID, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidEventID, err
	}

	eventId, err := sdktypes.Strict(sdktypes.ParseEventID(resp.Msg.EventId))
	if err != nil {
		return sdktypes.InvalidEventID, fmt.Errorf("invalid event id: %w", err)
	}

	return eventId, nil
}

func (c *client) Dispatch(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (sdktypes.EventID, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	resp, err := c.client.Dispatch(ctx, connect.NewRequest(&dispatcherv1.DispatchRequest{
		Event:        event.ToProto(),
		DeploymentId: opts.DeploymentID.String(),
		Env:          opts.Env,
	}))
	if err != nil {
		return sdktypes.InvalidEventID, rpcerrors.TranslateError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidEventID, err
	}

	eventId, err := sdktypes.Strict(sdktypes.ParseEventID(resp.Msg.EventId))
	if err != nil {
		return sdktypes.InvalidEventID, fmt.Errorf("invalid event id: %w", err)
	}

	return eventId, nil
}
