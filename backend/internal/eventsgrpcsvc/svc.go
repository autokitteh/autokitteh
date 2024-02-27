package eventsgrpcsvc

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

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

func Init(mux *http.ServeMux, events sdkservices.Events) {
	srv := server{events: events}

	path, namer := eventsv1connect.NewEventsServiceHandler(&srv)
	mux.Handle(path, namer)
}

func (s *server) Get(ctx context.Context, req *connect.Request[eventsv1.GetRequest]) (*connect.Response[eventsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eventId, err := sdktypes.StrictParseEventID(msg.EventId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	event, err := s.events.Get(ctx, eventId)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&eventsv1.GetResponse{Event: event.ToProto()}), nil
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

	filter := sdkservices.ListEventsFilter{
		IntegrationID:    iid,
		IntegrationToken: msg.IntegrationToken,
		OriginalID:       msg.OriginalId,
		EventType:        msg.EventType,
	}

	events, err := s.events.List(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	eventsPB := kittehs.Transform(events, sdktypes.ToProto)

	return connect.NewResponse(&eventsv1.ListResponse{Events: eventsPB}), nil
}

func (s *server) Save(ctx context.Context, req *connect.Request[eventsv1.SaveRequest]) (*connect.Response[eventsv1.SaveResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	event, err := sdktypes.EventFromProto(msg.Event)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	eid, err := s.events.Save(ctx, event)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&eventsv1.SaveResponse{EventId: eid.String()}), nil
}

func (s *server) AddEventRecord(ctx context.Context, req *connect.Request[eventsv1.AddEventRecordRequest]) (*connect.Response[eventsv1.AddEventRecordResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	record, err := sdktypes.EventRecordFromProto(msg.Record)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	err = s.events.AddEventRecord(ctx, record)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	return connect.NewResponse(&eventsv1.AddEventRecordResponse{}), nil
}

func (s *server) ListEventRecords(ctx context.Context, req *connect.Request[eventsv1.ListEventRecordsRequest]) (*connect.Response[eventsv1.ListEventRecordsResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eid, err := sdktypes.ParseEventID(msg.EventId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	filter := sdkservices.ListEventRecordsFilter{
		EventID: eid,
	}

	records, err := s.events.ListEventRecords(ctx, filter)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, fmt.Errorf("server error: %w", err))
	}

	recordsPB := kittehs.Transform(records, sdktypes.ToProto)

	return connect.NewResponse(&eventsv1.ListEventRecordsResponse{Records: recordsPB}), nil
}
