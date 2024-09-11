package sessionsgrpcsvc

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
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

func Init(muxes *muxes.Muxes, sessions sdkservices.Sessions) {
	srv := server{sessions: sessions}

	path, handler := sessionsv1connect.NewSessionsServiceHandler(&srv)
	muxes.Auth.Handle(path, handler)
}

func (s *server) Start(ctx context.Context, req *connect.Request[sessionsv1.StartRequest]) (*connect.Response[sessionsv1.StartResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	for k, v := range msg.JsonInputs {
		if msg.Session.Inputs == nil {
			msg.Session.Inputs = make(map[string]*sdktypes.ValuePB)
		}

		v, err := sdktypes.WrapValue(v)
		if err != nil {
			err = sdkerrors.NewInvalidArgumentError(`json_inputs["%s"]: %w`, k, err)
			return nil, sdkerrors.AsConnectError(fmt.Errorf(`json_inputs["%s"]: %w`, k, err))
		}

		msg.Session.Inputs[k] = v.ToProto()
	}

	session, err := sdktypes.SessionFromProto(msg.Session)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	uid, err := s.sessions.Start(ctx, session)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&sessionsv1.StartResponse{SessionId: uid.String()}), nil
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

	pb := session.ToProto()

	if req.Msg.JsonValues {
		if pb.Inputs, err = kittehs.TransformMapValuesError(pb.Inputs, sdktypes.ValueProtoToJSONStringValue); err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&sessionsv1.GetResponse{Session: pb}), nil
}

func (s *server) Stop(ctx context.Context, req *connect.Request[sessionsv1.StopRequest]) (*connect.Response[sessionsv1.StopResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	sessionID, err := sdktypes.ParseSessionID(msg.SessionId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	if err := s.sessions.Stop(ctx, sessionID, msg.Reason, msg.Terminate); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&sessionsv1.StopResponse{}), nil
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

	filter := sdkservices.ListSessionLogRecordsFilter{
		PaginationRequest: sdktypes.PaginationRequest{
			Skip:      msg.Skip,
			PageToken: msg.PageToken,
			PageSize:  msg.PageSize,
			Ascending: msg.Ascending,
		},
		SessionID: sessionID,
	}

	hist, err := s.sessions.GetLog(ctx, filter)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			return connect.NewResponse(&sessionsv1.GetLogResponse{}), nil
		}
		return nil, sdkerrors.AsConnectError(err)
	}

	logpb := hist.Log.ToProto()
	if req.Msg.JsonValues {
		for _, r := range logpb.Records {
			if c := r.CallAttemptComplete; c != nil {
				if c.Result.Value, err = sdktypes.ValueProtoToJSONStringValue(c.Result.Value); err != nil {
					return nil, err
				}
			} else if c := r.CallSpec; c != nil {
				if c.Args, err = kittehs.TransformError(c.Args, sdktypes.ValueProtoToJSONStringValue); err != nil {
					return nil, err
				}

				if c.Kwargs, err = kittehs.TransformMapValuesError(c.Kwargs, sdktypes.ValueProtoToJSONStringValue); err != nil {
					return nil, err
				}
			}
		}
	}

	return connect.NewResponse(&sessionsv1.GetLogResponse{Log: logpb, Count: hist.TotalCount, NextPageToken: hist.NextPageToken}), nil
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
		PaginationRequest: sdktypes.PaginationRequest{
			Skip:      msg.Skip,
			PageToken: msg.PageToken,
		},
	}

	filter.PageSize = msg.PageSize
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	if filter.PageSize < 10 {
		filter.PageSize = 10
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

	result, err := s.sessions.List(ctx, filter)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pbsessions := kittehs.Transform(result.Sessions, sdktypes.ToProto)

	return connect.NewResponse(&sessionsv1.ListResponse{Sessions: pbsessions, Count: result.TotalCount, NextPageToken: result.NextPageToken}), nil
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
