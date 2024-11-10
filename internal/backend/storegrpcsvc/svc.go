package storegrpcsvc

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	storev1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/store/v1/storev1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var errNotConfigured = connect.NewError(connect.CodeUnimplemented, fmt.Errorf("store service not configured"))

type server struct {
	store sdkservices.Store

	storev1connect.UnimplementedStoreServiceHandler
}

var _ storev1connect.StoreServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, store sdkservices.Store) {
	srv := server{store: store}

	path, handler := storev1connect.NewStoreServiceHandler(&srv)
	muxes.Auth.Handle(path, handler)
}

func (s *server) List(ctx context.Context, req *connect.Request[storev1.ListRequest]) (*connect.Response[storev1.ListResponse], error) {
	if s.store == nil {
		return nil, errNotConfigured
	}

	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	projectID, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	ks, err := s.store.List(ctx, projectID)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&storev1.ListResponse{Keys: ks}), nil
}

func (s *server) Get(ctx context.Context, req *connect.Request[storev1.GetRequest]) (*connect.Response[storev1.GetResponse], error) {
	if s.store == nil {
		return nil, errNotConfigured
	}

	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	projectID, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	vs, err := s.store.Get(ctx, projectID, msg.Keys)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&storev1.GetResponse{Values: kittehs.TransformMapValues(vs, sdktypes.ToProto)}), nil
}
