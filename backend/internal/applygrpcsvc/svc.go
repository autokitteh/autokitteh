package applygrpcsvc

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/manifest"
	"go.autokitteh.dev/autokitteh/proto"
	applyv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/apply/v1/applyv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type server struct {
	client sdkservices.Services
	applyv1connect.UnimplementedApplyServiceHandler
}

var _ applyv1connect.ApplyServiceHandler = (*server)(nil)

func Init(mux *http.ServeMux, client sdkservices.Services) {
	srv := server{client: client}

	path, namer := applyv1connect.NewApplyServiceHandler(&srv)
	mux.Handle(path, namer)
}

func (s *server) Apply(ctx context.Context, req *connect.Request[applyv1.ApplyRequest]) (*connect.Response[applyv1.ApplyResponse], error) {
	msg := req.Msg

	if err := proto.Validate(msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	man, err := manifest.Read([]byte(msg.Manifest), msg.Path)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var logs []string

	actions, err := manifest.Plan(ctx, man, s.client, manifest.WithLogger(func(msg string) {
		logs = append(logs, fmt.Sprintf("[plan] %s", msg))
	}))
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	pids, err := manifest.Execute(ctx, actions, s.client, func(msg string) {
		logs = append(logs, fmt.Sprintf("[exec] %s", msg))
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	stringPIDs := kittehs.Transform(pids, func(pid sdktypes.ProjectID) string { return pid.String() })
	return connect.NewResponse(&applyv1.ApplyResponse{Logs: logs, ProjectIds: stringPIDs}), nil
}
