package sdkruntimessvc

import (
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
		return nil, fmt.Errorf("ls: %w", err)
	}

	return s.ls, nil
}

func (s *svc) Describe(ctx context.Context, req *connect.Request[runtimesv1.DescribeRequest]) (*connect.Response[runtimesv1.DescribeResponse], error) {
	err := akproto.Validate(req.Msg)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrRPC, err)
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
		return nil, fmt.Errorf("%w: %v", sdkerrors.ErrRPC, err)
	}

	rts, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimesv1.ListResponse{Runtimes: kittehs.Transform(rts, sdktypes.ToProto)}), nil
}
