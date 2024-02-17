package applygrpcsvc

import (
	"context"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/proto"
	applyv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1/applyv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type server struct {
	apply sdkservices.Apply
	applyv1connect.UnimplementedApplyServiceHandler
}

var _ applyv1connect.ApplyServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, apply sdkservices.Apply) {
	srv := server{apply: apply}

	path, namer := applyv1connect.NewApplyServiceHandler(&srv)
	mux.Handle(path, namer)
}

func (s *server) Apply(ctx context.Context, req *connect.Request[applyv1.ApplyRequest]) (*connect.Response[applyv1.ApplyResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	logs, err := s.apply.Apply(ctx, msg.Manifest, msg.Path)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&applyv1.ApplyResponse{Logs: logs}), nil
}

func (s *server) Plan(ctx context.Context, req *connect.Request[applyv1.PlanRequest]) (*connect.Response[applyv1.PlanResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	logs, err := s.apply.Plan(ctx, msg.Manifest)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return connect.NewResponse(&applyv1.PlanResponse{Logs: logs}), nil
}
