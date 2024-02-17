package sdkruntimessvc

import (
	"context"
	"fmt"
	"net/url"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/proto"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (s *svc) Build(ctx context.Context, req *connect.Request[runtimesv1.BuildRequest]) (*connect.Response[runtimesv1.BuildResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	name, err := sdktypes.StrictParseName(msg.RuntimeName)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("name: %w", err))
	}

	syms, err := kittehs.TransformError(msg.ValueNames, sdktypes.StrictParseSymbol)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("value_names: %w", err))
	}

	rootURL, err := url.Parse(msg.RootUrl)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("url: %w", err))
	}

	rt, err := s.runtimes.New(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("new: %w", err)
	}

	a, err := rt.Build(ctx, rootURL, msg.Path, syms)
	if err != nil {
		if perr := sdktypes.ProgramErrorFromError(err); perr != nil {
			return connect.NewResponse(&runtimesv1.BuildResponse{
				Error: perr.ToProto(),
			}), nil
		}
		return nil, err
	}

	return connect.NewResponse(&runtimesv1.BuildResponse{
		Product: a.ToProto(),
	}), nil
}
