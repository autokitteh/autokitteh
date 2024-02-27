package envsgrpcsvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	envsv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/envs/v1/envsv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	envs sdkservices.Envs

	envsv1connect.UnimplementedEnvsServiceHandler
}

var _ envsv1connect.EnvsServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, envs sdkservices.Envs) {
	srv := server{envs: envs}

	path, handler := envsv1connect.NewEnvsServiceHandler(&srv)
	mux.Handle(path, handler)
}

func (s *server) Create(ctx context.Context, req *connect.Request[envsv1.CreateRequest]) (*connect.Response[envsv1.CreateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	env, err := sdktypes.EnvFromProto(msg.Env)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	eid, err := s.envs.Create(ctx, env)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&envsv1.CreateResponse{EnvId: eid.String()}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[envsv1.GetRequest]) (*connect.Response[envsv1.GetResponse], error) {
	toResponse := func(env sdktypes.Env, err error) (*connect.Response[envsv1.GetResponse], error) {
		if err != nil {
			if errors.Is(err, sdkerrors.ErrNotFound) {
				return connect.NewResponse(&envsv1.GetResponse{}), nil
			} else if errors.Is(err, sdkerrors.ErrUnauthorized) {
				return nil, connect.NewError(connect.CodePermissionDenied, err)
			}

			return nil, connect.NewError(connect.CodeUnknown, err)
		}

		return connect.NewResponse(&envsv1.GetResponse{Env: env.ToProto()}), nil
	}

	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eid, err := sdktypes.ParseEnvID(msg.EnvId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("env_id: %w", err))
	}

	if eid != nil {
		return toResponse(s.envs.GetByID(ctx, eid))
	}

	// a handle must've been supplied here.
	n, err := sdktypes.StrictParseName(msg.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name: %w", err))
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("project_id: %w", err))
	}

	return toResponse(s.envs.GetByName(ctx, pid, n))
}

func (s *server) List(ctx context.Context, req *connect.Request[envsv1.ListRequest]) (*connect.Response[envsv1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(req.Msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("project_id: %w", err))
	}

	envs, err := s.envs.List(ctx, pid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pbenvs := kittehs.Transform(envs, sdktypes.ToProto)

	return connect.NewResponse(&envsv1.ListResponse{Envs: pbenvs}), nil
}

func (s *server) SetVar(ctx context.Context, req *connect.Request[envsv1.SetVarRequest]) (*connect.Response[envsv1.SetVarResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ev, err := sdktypes.StrictEnvVarFromProto(req.Msg.Var)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("var: %w", err))
	}

	if err := s.envs.SetVar(ctx, ev); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&envsv1.SetVarResponse{}), nil
}

func (s *server) RevealVar(ctx context.Context, req *connect.Request[envsv1.RevealVarRequest]) (*connect.Response[envsv1.RevealVarResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eid, err := sdktypes.StrictParseEnvID(req.Msg.EnvId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("env_id: %w", err))
	}

	vn, err := sdktypes.ParseSymbol(req.Msg.Name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name: %w", err))
	}

	v, err := s.envs.RevealVar(ctx, eid, vn)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&envsv1.RevealVarResponse{Value: v}), nil
}

func (s *server) GetVars(ctx context.Context, req *connect.Request[envsv1.GetVarsRequest]) (*connect.Response[envsv1.GetVarsResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	eid, err := sdktypes.StrictParseEnvID(req.Msg.EnvId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("owner_id: %w", err))
	}

	vns, err := kittehs.TransformError(req.Msg.Names, sdktypes.ParseSymbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("names: %w", err))
	}

	evs, err := s.envs.GetVars(ctx, vns, eid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pbevs := kittehs.Transform(evs, sdktypes.ToProto)

	return connect.NewResponse(&envsv1.GetVarsResponse{Vars: pbevs}), nil
}
