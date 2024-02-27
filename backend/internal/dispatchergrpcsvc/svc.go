package dispatchergrpcsvc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/backend/internal/dispatcher"
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

func Init(mux *http.ServeMux, dispatcher dispatcher.Dispatcher) {
	srv := server{dispatcher: dispatcher}

	path, namer := dispatcherv1connect.NewDispatcherServiceHandler(&srv)
	mux.Handle(path, namer)
}

func (s *server) Dispatch(ctx context.Context, req *connect.Request[dispatcher1.DispatchRequest]) (*connect.Response[dispatcher1.DispatchResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	event, err := sdktypes.EventFromProto(msg.Event)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	envID, err := sdktypes.ParseEnvID(req.Msg.EnvId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	deploymentID, err := sdktypes.ParseDeploymentID(req.Msg.DeploymentId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eventID, err := s.dispatcher.Dispatch(ctx, event, &sdkservices.DispatchOptions{EnvID: envID, DeploymentID: deploymentID})
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

	eventID, err := sdktypes.StrictParseEventID(req.Msg.EventId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	envID, err := sdktypes.ParseEnvID(req.Msg.EnvId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	deploymentID, err := sdktypes.ParseDeploymentID(req.Msg.DeploymentId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	newEventID, err := s.dispatcher.Redispatch(ctx, eventID, &sdkservices.DispatchOptions{EnvID: envID, DeploymentID: deploymentID})
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&dispatcher1.RedispatchResponse{EventId: newEventID.String()}), nil
}
