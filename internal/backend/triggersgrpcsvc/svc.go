package triggersgrpcsvc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	triggersv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/triggers/v1/triggersv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	triggers sdkservices.Triggers

	triggersv1connect.UnimplementedTriggersServiceHandler
}

var _ triggersv1connect.TriggersServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, triggers sdkservices.Triggers) {
	srv := server{triggers: triggers}

	path, namer := triggersv1connect.NewTriggersServiceHandler(&srv)
	muxes.Main.Auth.Handle(path, namer)
}

func (s *server) Create(ctx context.Context, req *connect.Request[triggersv1.CreateRequest]) (*connect.Response[triggersv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	trigger, err := sdktypes.StrictTriggerFromProto(msg.Trigger)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	mid, err := s.triggers.Create(ctx, trigger)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&triggersv1.CreateResponse{TriggerId: mid.String()}), nil
}

func (s *server) Update(ctx context.Context, req *connect.Request[triggersv1.UpdateRequest]) (*connect.Response[triggersv1.UpdateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	trigger, err := sdktypes.StrictTriggerFromProto(msg.Trigger)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.triggers.Update(ctx, trigger); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&triggersv1.UpdateResponse{}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[triggersv1.DeleteRequest]) (*connect.Response[triggersv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	mid, err := sdktypes.ParseTriggerID(msg.TriggerId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.triggers.Delete(ctx, mid); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&triggersv1.DeleteResponse{}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[triggersv1.GetRequest]) (*connect.Response[triggersv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	mid, err := sdktypes.ParseTriggerID(msg.TriggerId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	trigger, err := s.triggers.Get(ctx, mid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&triggersv1.GetResponse{Trigger: trigger.ToProto()}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[triggersv1.ListRequest]) (*connect.Response[triggersv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eid, err := sdktypes.ParseEnvID(msg.EnvId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	cid, err := sdktypes.ParseConnectionID(msg.ConnectionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	stype, err := sdktypes.TriggerSourceTypeFromProto(msg.SourceType)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	filter := sdkservices.ListTriggersFilter{
		EnvID:        eid,
		ConnectionID: cid,
		ProjectID:    pid,
		SourceType:   stype,
	}

	triggers, err := s.triggers.List(ctx, filter)
	if err != nil {
		return nil, sdkerrors.AsConnectError(fmt.Errorf("server error: %w", err))
	}

	triggersPB := kittehs.Transform(triggers, sdktypes.ToProto)
	return connect.NewResponse(&triggersv1.ListResponse{Triggers: triggersPB}), nil
}
