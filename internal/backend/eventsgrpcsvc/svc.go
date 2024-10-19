package eventsgrpcsvc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	eventsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/events/v1/eventsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	events sdkservices.Events

	eventsv1connect.UnimplementedEventsServiceHandler
}

var _ eventsv1connect.EventsServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, events sdkservices.Events) {
	srv := server{events: events}

	path, namer := eventsv1connect.NewEventsServiceHandler(&srv)
	muxes.Main.Auth.Handle(path, namer)
}

func (s *server) Get(ctx context.Context, req *connect.Request[eventsv1.GetRequest]) (*connect.Response[eventsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eventId, err := sdktypes.Strict(sdktypes.ParseEventID(msg.EventId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	event, err := s.events.Get(ctx, eventId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eventpb := event.ToProto()

	if msg.JsonValues {
		if eventpb.Data, err = kittehs.TransformMapValuesError(eventpb.Data, sdktypes.ValueProtoToJSONStringValue); err != nil {
			return nil, sdkerrors.AsConnectError(err)
		}
	}

	return connect.NewResponse(&eventsv1.GetResponse{Event: eventpb}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[eventsv1.ListRequest]) (*connect.Response[eventsv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	iid, err := sdktypes.ParseIntegrationID(msg.IntegrationId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	did, err := sdktypes.ParseEventDestinationID(msg.DestinationId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	order := sdkservices.ListOrder(msg.Order)

	// set default order if not set
	if msg.Order == "" {
		order = sdkservices.ListOrderDescending
	}

	// verify order is valid
	if order != sdkservices.ListOrderAscending && order != sdkservices.ListOrderDescending {
		err := fmt.Errorf("order should be either %s or %s", sdkservices.ListOrderAscending, sdkservices.ListOrderDescending)
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	filter := sdkservices.ListEventsFilter{
		IntegrationID: iid,
		DestinationID: did,
		EventType:     msg.EventType,
		Limit:         int(msg.MaxResults),
		Order:         order,
	}

	events, err := s.events.List(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	eventspb := kittehs.Transform(events, sdktypes.ToProto)

	if msg.JsonValues {
		for _, eventpb := range eventspb {
			if eventpb.Data, err = kittehs.TransformMapValuesError(eventpb.Data, sdktypes.ValueProtoToJSONStringValue); err != nil {
				return nil, sdkerrors.AsConnectError(err)
			}
		}
	}

	return connect.NewResponse(&eventsv1.ListResponse{Events: eventspb}), nil
}

func (s *server) Save(ctx context.Context, req *connect.Request[eventsv1.SaveRequest]) (*connect.Response[eventsv1.SaveResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	event, err := sdktypes.EventFromProto(msg.Event)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eid, err := s.events.Save(ctx, event)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&eventsv1.SaveResponse{EventId: eid.String()}), nil
}
