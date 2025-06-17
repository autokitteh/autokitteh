package storegrpcsvc

import (
	"context"

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

type server struct{ store sdkservices.Store }

var _ storev1connect.StoreServiceHandler = (*server)(nil)

func Init(muxes *muxes.Muxes, store sdkservices.Store) {
	s := &server{store: store}
	path, namer := storev1connect.NewStoreServiceHandler(s)
	muxes.Auth.Handle(path, namer)
}

func (s *server) Get(ctx context.Context, req *connect.Request[storev1.GetRequest]) (*connect.Response[storev1.GetResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	vs, err := s.store.Get(ctx, pid, msg.Keys)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&storev1.GetResponse{
		Values: kittehs.TransformMapValues(vs, sdktypes.ToProto),
	}), nil
}

func (s *server) List(ctx context.Context, req *connect.Request[storev1.ListRequest]) (*connect.Response[storev1.ListResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	keys, err := s.store.List(ctx, pid)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&storev1.ListResponse{Keys: keys}), nil
}

func (s *server) Mutate(ctx context.Context, req *connect.Request[storev1.MutateRequest]) (*connect.Response[storev1.MutateResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	pid, err := sdktypes.ParseProjectID(msg.ProjectId)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	operands, err := kittehs.TransformError(msg.Operands, sdktypes.ValueFromProto)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	v, err := s.store.Mutate(ctx, pid, msg.Key, msg.Operation, operands...)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	return connect.NewResponse(&storev1.MutateResponse{Value: v.ToProto()}), nil
}
