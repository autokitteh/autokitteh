package sdkruntimesclient

import (
	"bytes"
	"context"
	"fmt"

	"connectrpc.com/connect"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/sdk/internal/rpcerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type run struct {
	stream *connect.ServerStreamForClient[runtimesv1.RunResponse]
	rid    sdktypes.RunID
	result map[string]sdktypes.Value
}

func (r *run) ID() sdktypes.RunID                { return r.rid }
func (r *run) Values() map[string]sdktypes.Value { return r.result }
func (r *run) ExecutorID() sdktypes.ExecutorID   { return sdktypes.NewExecutorID(r.rid) }
func (r *run) Close()                            { r.stream.Close() }

func (r *run) Call(context.Context, sdktypes.Value, []sdktypes.Value, map[string]sdktypes.Value) (sdktypes.Value, error) {
	// Need a way to pass this to the server - will do when the stream will be bidi.
	return sdktypes.InvalidValue, sdkerrors.ErrNotImplemented
}

func (c *client) run(
	ctx context.Context,
	rid sdktypes.RunID,
	path string,
	build *sdkbuildfile.BuildFile,
	globals map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	var a bytes.Buffer
	if err := build.Write(&a); err != nil {
		return nil, fmt.Errorf("failed to write build file: %w", err)
	}

	stream, err := c.client.Run(ctx, connect.NewRequest(&runtimesv1.RunRequest{
		RunId:    rid.String(),
		Path:     path,
		Globals:  kittehs.TransformMapValues(globals, sdktypes.ToProto),
		Artifact: a.Bytes(),
	}))
	if err != nil {
		return nil, rpcerrors.ToSDKError(err)
	}

	var result map[string]sdktypes.Value

	for stream.Receive() {
		msg := stream.Msg()
		if msg.Error != nil {
			stream.Close()
			perr, err := sdktypes.ProgramErrorFromProto(msg.Error)
			if err != nil {
				return nil, fmt.Errorf("invalid error: %w", err)
			}
			return nil, perr.ToError()
		}

		if msg.Print != "" {
			cbs.Print(ctx, rid, msg.Print)
		}

		if msg.Result != nil {
			result, err = kittehs.TransformMapValuesError(msg.Result, sdktypes.StrictValueFromProto)
			if err != nil {
				stream.Close()
				return nil, fmt.Errorf("invalid result: %w", err)
			}
		}
	}

	// TODO: in the future when we'll support calling run functions
	//       we'll pass the stream to the run object and let it handle it.
	return &run{stream: stream, rid: rid, result: result}, nil
}
