package envsgrpcsvc

import (
	"context"
	"errors"
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
		return nil, sdkerrors.AsConnectError(err)
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
		return nil, sdkerrors.AsConnectError(err)
	}

	if eid.IsValid() {
		return toResponse(s.envs.GetByID(ctx, eid))
	}

	// a handle must've been supplied here.
	n, err := sdktypes.StrictParseSymbol(msg.Name)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
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
		return nil, sdkerrors.AsConnectError(err)
	}

	envs, err := s.envs.List(ctx, pid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pbenvs := kittehs.Transform(envs, sdktypes.ToProto)

	return connect.NewResponse(&envsv1.ListResponse{Envs: pbenvs}), nil
}
