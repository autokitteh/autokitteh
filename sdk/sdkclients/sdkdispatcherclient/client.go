package sdkdispatcherclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
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

func (c *client) Redispatch(ctx context.Context, eventID sdktypes.EventID, opts *sdkservices.DispatchOptions) (*sdkservices.DispatchResponse, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	resp, err := c.client.Redispatch(
		ctx,
		connect.NewRequest(
			&dispatcherv1.RedispatchRequest{
				EventId:      eventID.String(),
				DeploymentId: opts.DeploymentID.String(),
				Wait:         opts.Wait,
			},
		),
	)
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	eventId, err := sdktypes.Strict(sdktypes.ParseEventID(resp.Msg.EventId))
	if err != nil {
		return nil, fmt.Errorf("invalid event id: %w", err)
	}

	sids, err := kittehs.TransformError(resp.Msg.SessionIds, sdktypes.ParseSessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session id: %w", err)
	}

	return &sdkservices.DispatchResponse{
		EventID:    eventId,
		SessionIDs: sids,
	}, nil
}

func (c *client) Dispatch(ctx context.Context, event sdktypes.Event, opts *sdkservices.DispatchOptions) (*sdkservices.DispatchResponse, error) {
	if opts == nil {
		opts = &sdkservices.DispatchOptions{}
	}

	resp, err := c.client.Dispatch(ctx, connect.NewRequest(&dispatcherv1.DispatchRequest{
		Event:        event.ToProto(),
		DeploymentId: opts.DeploymentID.String(),
		Wait:         opts.Wait,
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	eventId, err := sdktypes.Strict(sdktypes.ParseEventID(resp.Msg.EventId))
	if err != nil {
		return nil, fmt.Errorf("invalid event id: %w", err)
	}

	sids, err := kittehs.TransformError(resp.Msg.SessionIds, sdktypes.ParseSessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session id: %w", err)
	}

	return &sdkservices.DispatchResponse{
		EventID:    eventId,
		SessionIDs: sids,
	}, nil
}
