package sdkruntimessvc

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	akproto "go.autokitteh.dev/autokitteh/proto"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1/runtimesv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type svc struct {
	runtimes sdkservices.Runtimes

	lsOnce sync.Once
	ls     []sdktypes.Runtime

	runtimesv1connect.UnimplementedRuntimesServiceHandler
}

var _ runtimesv1connect.RuntimesServiceHandler = &svc{}

func Init(runtimes sdkservices.Runtimes, mux *http.ServeMux) {
	path, h := runtimesv1connect.NewRuntimesServiceHandler(&svc{runtimes: runtimes})
	mux.Handle(path, h)
}

func (s *svc) list(ctx context.Context) ([]sdktypes.Runtime, error) {
	var err error

	// Ugly, but works for now.
	s.lsOnce.Do(func() { s.ls, err = s.runtimes.List(ctx) })

	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return s.ls, nil
}

func (s *svc) Describe(ctx context.Context, req *connect.Request[runtimesv1.DescribeRequest]) (*connect.Response[runtimesv1.DescribeResponse], error) {
	err := akproto.Validate(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	rts, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	for _, r := range rts {
		if r.ToProto().Name == req.Msg.Name {
			return connect.NewResponse(&runtimesv1.DescribeResponse{Runtime: r.ToProto()}), nil
		}
	}

	return connect.NewResponse(&runtimesv1.DescribeResponse{}), nil
}

func (s *svc) List(ctx context.Context, req *connect.Request[runtimesv1.ListRequest]) (*connect.Response[runtimesv1.ListResponse], error) {
	if err := akproto.Validate(req.Msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	rts, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimesv1.ListResponse{Runtimes: kittehs.Transform(rts, sdktypes.ToProto)}), nil
}

func (s *svc) Build(ctx context.Context, req *connect.Request[runtimesv1.BuildRequest]) (*connect.Response[runtimesv1.BuildResponse], error) {
	if err := akproto.Validate(req.Msg); err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrRPC, err)
	}

	symbols, err := kittehs.TransformError(req.Msg.Symbols, sdktypes.StrictParseSymbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	srcFS, err := kittehs.MapToMemFS(req.Msg.Resources)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	bf, err := s.runtimes.Build(ctx, srcFS, symbols, req.Msg.Memo)
	if err != nil {
		if err := sdktypes.ProgramErrorFromError(err); err != nil {
			return connect.NewResponse(&runtimesv1.BuildResponse{Error: err.ToProto()}), nil
		}
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	var buf bytes.Buffer
	if err := bf.Write(&buf); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&runtimesv1.BuildResponse{
		Artifact: buf.Bytes(),
	}), nil
}
