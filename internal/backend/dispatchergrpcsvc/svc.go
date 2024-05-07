package dispatchergrpcsvc

import (
	"context"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/dispatcher"
	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/proto"
	dispatcher1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/dispatcher/v1/dispatcherv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	dispatcher dispatcher.Dispatcher

	dispatcherv1connect.UnimplementedDispatcherServiceHandler
}

var _ dispatcherv1connect.DispatcherServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, dispatcher dispatcher.Dispatcher) {
	srv := server{dispatcher: dispatcher}

	path, namer := dispatcherv1connect.NewDispatcherServiceHandler(&srv)
	muxes.Auth.Handle(path, namer)
}

func (s *server) Dispatch(ctx context.Context, req *connect.Request[dispatcher1.DispatchRequest]) (*connect.Response[dispatcher1.DispatchResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	event, err := sdktypes.EventFromProto(msg.Event)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	deploymentID, err := sdktypes.ParseDeploymentID(msg.DeploymentId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eventID, err := s.dispatcher.Dispatch(ctx, event, &sdkservices.DispatchOptions{Env: msg.Env, DeploymentID: deploymentID})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&dispatcher1.DispatchResponse{EventId: eventID.String()}), nil
}

func (s *server) Redispatch(ctx context.Context, req *connect.Request[dispatcher1.RedispatchRequest]) (*connect.Response[dispatcher1.RedispatchResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eventID, err := sdktypes.Strict(sdktypes.ParseEventID(msg.EventId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	deploymentID, err := sdktypes.Strict(sdktypes.ParseDeploymentID(msg.DeploymentId))
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	newEventID, err := s.dispatcher.Redispatch(ctx, eventID, &sdkservices.DispatchOptions{Env: msg.EnvId, DeploymentID: deploymentID})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&dispatcher1.RedispatchResponse{EventId: newEventID.String()}), nil
}
