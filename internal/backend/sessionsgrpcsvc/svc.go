package sessionsgrpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	sessionsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/sessions/v1/sessionsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	sessions sdkservices.Sessions

	sessionsv1connect.UnimplementedSessionsServiceHandler
}

var _ sessionsv1connect.SessionsServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, sessions sdkservices.Sessions) {
	srv := server{sessions: sessions}

	path, handler := sessionsv1connect.NewSessionsServiceHandler(&srv)
	mux.Handle(path, handler)
}

func (s *server) Start(ctx context.Context, req *connect.Request[sessionsv1.StartRequest]) (*connect.Response[sessionsv1.StartResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	session, err := sdktypes.SessionFromProto(msg.Session)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sid, err := s.sessions.Start(ctx, session)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&sessionsv1.StartResponse{SessionId: sid.String()}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[sessionsv1.GetRequest]) (*connect.Response[sessionsv1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	session, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&sessionsv1.GetResponse{Session: session.ToProto()}), nil
}

func (s *server) GetLog(ctx context.Context, req *connect.Request[sessionsv1.GetLogRequest]) (*connect.Response[sessionsv1.GetLogResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	hist, err := s.sessions.GetLog(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return connect.NewResponse(&sessionsv1.GetLogResponse{}), nil
		}
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&sessionsv1.GetLogResponse{Log: hist.ToProto()}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[sessionsv1.ListRequest]) (*connect.Response[sessionsv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	stateType, err := sdktypes.SessionStateTypeFromProto(msg.StateType)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("state_type: %w", err))
	}

	filter := sdkservices.ListSessionsFilter{
		StateType: stateType,
		CountOnly: msg.CountOnly,
	}

	if filter.DeploymentID, err = sdktypes.ParseDeploymentID(req.Msg.DeploymentId); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if filter.EventID, err = sdktypes.ParseEventID(req.Msg.EventId); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if filter.EnvID, err = sdktypes.ParseEnvID(req.Msg.EnvId); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessions, n, err := s.sessions.List(ctx, filter)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pbsessions := kittehs.Transform(sessions, sdktypes.ToProto)

	return connect.NewResponse(&sessionsv1.ListResponse{Sessions: pbsessions, Count: int32(n)}), nil
}

func (s *server) Delete(ctx context.Context, req *connect.Request[sessionsv1.DeleteRequest]) (*connect.Response[sessionsv1.DeleteResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err = s.sessions.Delete(ctx, sessionID); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&sessionsv1.DeleteResponse{}), nil
}

func (s *server) Stop(ctx context.Context, req *connect.Request[sessionsv1.StopRequest]) (*connect.Response[sessionsv1.StopResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.StrictParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err = s.sessions.Stop(ctx, sessionID, msg.Reason, msg.Terminate); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}
	return connect.NewResponse(&sessionsv1.StopResponse{}), nil
}
