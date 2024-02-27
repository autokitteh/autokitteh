package storegrpcsvc

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	storev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1/storev1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	store sdkservices.Store

	storev1connect.UnimplementedStoreServiceHandler
}

var _ storev1connect.StoreServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, store sdkservices.Store) {
	srv := server{store: store}

	path, handler := storev1connect.NewStoreServiceHandler(&srv)
	mux.Handle(path, handler)
}

func (s *server) List(ctx context.Context, req *connect.Request[storev1.ListRequest]) (*connect.Response[storev1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	envID, err := sdktypes.ParseEnvID(msg.EnvId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("env_id: %w", err))
	}

	projectID, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("project_id: %w", err))
	}

	ks, err := s.store.List(ctx, envID, projectID)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&storev1.ListResponse{Keys: ks}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[storev1.GetRequest]) (*connect.Response[storev1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	envID, err := sdktypes.ParseEnvID(msg.EnvId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("env_id: %w", err))
	}

	projectID, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("project_id: %w", err))
	}

	vs, err := s.store.Get(ctx, envID, projectID, msg.Keys)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&storev1.GetResponse{Values: kittehs.TransformMapValues(vs, sdktypes.ToProto)}), nil
}
