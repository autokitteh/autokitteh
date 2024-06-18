package sdkruntimesclient

import (
	"context"

	"connectrpc.com/connect"

	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type run struct {
	stream *connect.ServerStreamForClient[runtimesv1.RunResponse]
	rid    sdktypes.RunID
	result map[string]sdktypes.Value
}

func (r *run) ID() sdktypes.RunID                { return r.rid }
func (r *run) Values() map[string]sdktypes.Value { return r.result }
func (r *run) ExecutorIDs() []sdktypes.ExecutorID {
	return []sdktypes.ExecutorID{sdktypes.NewExecutorID(r.rid)}
}
func (r *run) Close() { r.stream.Close() }

func (r *run) Call(context.Context, sdktypes.Value, []sdktypes.Value, map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Need a way to pass this to the server - will do when the stream will be bidi.
	return sdktypes.InvalidValue, sdkerrors.ErrNotImplemented
}
