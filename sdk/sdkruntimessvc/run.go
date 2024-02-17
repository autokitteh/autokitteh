package sdkruntimessvc

import (
	"context"

	"connectrpc.com/connect"

	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
)

func (s *svc) Run(ctx context.Context, stream *connect.BidiStream[runtimesv1.RunRequest, runtimesv1.RunResponse]) error {
	return sdkerrors.ErrNotImplemented
}
