package sdkeventsclient

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1/eventsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/internal"
	"go.autokitteh.dev/autokitteh/sdk/sdkclients/sdkclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type client struct {
	client eventsv1connect.EventsServiceClient
}

func New(p sdkclient.Params) sdkservices.Events {
	return &client{client: internal.New(eventsv1connect.NewEventsServiceClient, p)}
}

func (c *client) Save(ctx context.Context, event sdktypes.Event) (sdktypes.EventID, error) {
	resp, err := c.client.Save(ctx, connect.NewRequest(&eventsv1.SaveRequest{Event: event.ToProto()}))
	if err != nil {
		return sdktypes.InvalidEventID, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidEventID, err
	}

	eventId, err := sdktypes.Strict(sdktypes.ParseEventID(resp.Msg.EventId))
	if err != nil {
		return sdktypes.InvalidEventID, fmt.Errorf("invalid event ID: %w", err)
	}

	return eventId, nil
}

func (c *client) Get(ctx context.Context, eventId sdktypes.EventID) (sdktypes.Event, error) {
	resp, err := c.client.Get(ctx, connect.NewRequest(
		&eventsv1.GetRequest{EventId: eventId.String()},
	))
	if err != nil {
		return sdktypes.InvalidEvent, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return sdktypes.InvalidEvent, err
	}

	event, err := sdktypes.EventFromProto(resp.Msg.Event)
	if err != nil {
		return sdktypes.InvalidEvent, fmt.Errorf("invalid event: %w", err)
	}
	return event, nil
}

func (c *client) List(ctx context.Context, filter sdkservices.ListEventsFilter) ([]sdktypes.Event, error) {
	resp, err := c.client.List(ctx, connect.NewRequest(
		&eventsv1.ListRequest{
			IntegrationId: filter.IntegrationID.String(),
			EventType:     filter.EventType,
			DestinationId: filter.DestinationID.String(),
			MaxResults:    uint32(filter.Limit),
			Order:         string(filter.Order),
		},
	))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	if err := internal.Validate(resp.Msg); err != nil {
		return nil, err
	}

	return kittehs.TransformError(resp.Msg.Events, sdktypes.StrictEventFromProto)
}
