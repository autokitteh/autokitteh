package sdkruntimesclient

import (
	"bytes"
	"context"
	"fmt"

	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/sdk/internal/bidirunmsgs"
	"go.autokitteh.dev/autokitteh/sdk/internal/loggedstream"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type bidiRun struct {
	stream loggedstream.Stream[runtimesv1.BidiRunResponse, runtimesv1.BidiRunRequest]
	rid    sdktypes.RunID
	values map[string]sdktypes.Value
	cbs    *sdkservices.RunCallbacks
}

func (r *bidiRun) ID() sdktypes.RunID                { return r.rid }
func (r *bidiRun) Values() map[string]sdktypes.Value { return r.values }
func (r *bidiRun) ExecutorID() sdktypes.ExecutorID   { return sdktypes.NewExecutorID(r.rid) }
func (r *bidiRun) Close() {
	// TODO
}

func (r *bidiRun) Call(ctx context.Context, fv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if err := r.stream.Send(&runtimesv1.BidiRunRequest{
		Request: &runtimesv1.BidiRunRequest_Call_{
			Call: &runtimesv1.BidiRunCall{
				Value:  fv.ToProto(),
				Args:   kittehs.Transform(args, sdktypes.ToProto),
				Kwargs: kittehs.TransformMapValues(kwargs, sdktypes.ToProto),
			},
		},
	}); err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("send call: %w", err)
	}

	var callret *runtimesv1.BidiRunCallReturn
	for {
		msg, err := r.stream.Receive()
		if err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("receive: %w", err)
		}

		if msg.GetCallReturn() != nil {
			callret = msg.GetCallReturn()
			break
		}

		if err := r.handleCallback(ctx, msg); err != nil {
			return sdktypes.InvalidValue, fmt.Errorf("callback: %w", err)
		}
	}

	if pb := callret.GetError(); pb != nil {
		perr, err := sdktypes.ProgramErrorFromProto(pb)
		if perr.IsValid() {
			err = perr.ToError()
		} else {
			err = fmt.Errorf("decode error: %w", err)
		}
		return sdktypes.InvalidValue, err
	}

	return sdktypes.StrictValueFromProto(callret.GetValue())
}

func (r *bidiRun) handleCallback(ctx context.Context, msg *runtimesv1.BidiRunResponse) error {
	switch msg := msg.Response.(type) {
	case *runtimesv1.BidiRunResponse_Print_:
		r.cbs.SafePrint(ctx, r.rid, msg.Print.Text)
		return nil

	case *runtimesv1.BidiRunResponse_NewRunId:
		rid := r.cbs.SafeNewRunID()

		if err := r.stream.Send(&runtimesv1.BidiRunRequest{
			Request: &runtimesv1.BidiRunRequest_NewRunIdValue{
				NewRunIdValue: &runtimesv1.BidiRunRequest_NewRunIDValue{
					RunId: rid.String(),
				},
			},
		}); err != nil {
			return fmt.Errorf("send new run id return: %w", err)
		}

		return nil

	case *runtimesv1.BidiRunResponse_Call:
		v, err := sdktypes.StrictValueFromProto(msg.Call.Value)
		if err != nil {
			return fmt.Errorf("decode value: %w", err)
		}

		args, err := kittehs.TransformError(msg.Call.Args, sdktypes.StrictValueFromProto)
		if err != nil {
			return fmt.Errorf("decode args: %w", err)
		}

		kwargs, err := kittehs.TransformMapValuesError(msg.Call.Kwargs, sdktypes.StrictValueFromProto)
		if err != nil {
			return fmt.Errorf("decode kwargs: %w", err)
		}

		ret, err := r.cbs.SafeCall(ctx, r.rid, v, args, kwargs)
		if err != nil {
			if sendErr := r.stream.Send(&runtimesv1.BidiRunRequest{
				Request: &runtimesv1.BidiRunRequest_CallReturn{
					CallReturn: &runtimesv1.BidiRunCallReturn{
						Result: &runtimesv1.BidiRunCallReturn_Error{
							Error: sdktypes.WrapError(err).ToProto(),
						},
					},
				},
			}); sendErr != nil {
				return fmt.Errorf("send call return: %w", err)
			}

			return nil
		}

		if err := r.stream.Send(&runtimesv1.BidiRunRequest{
			Request: &runtimesv1.BidiRunRequest_CallReturn{
				CallReturn: &runtimesv1.BidiRunCallReturn{
					Result: &runtimesv1.BidiRunCallReturn_Value{
						Value: ret.ToProto(),
					},
				},
			},
		}); err != nil {
			return fmt.Errorf("send call return: %w", err)
		}

		return nil

	case *runtimesv1.BidiRunResponse_Load_:
		path := msg.Load.Path

		vs, err := r.cbs.SafeLoad(ctx, r.rid, path)
		if err != nil {
			if err := r.stream.Send(&runtimesv1.BidiRunRequest{
				Request: &runtimesv1.BidiRunRequest_LoadReturn{
					LoadReturn: &runtimesv1.BidiRunLoadReturn{
						Error: sdktypes.WrapError(err).ToProto(),
					},
				},
			}); err != nil {
				return fmt.Errorf("send load return: %w", err)
			}
		}

		if err := r.stream.Send(&runtimesv1.BidiRunRequest{
			Request: &runtimesv1.BidiRunRequest_LoadReturn{
				LoadReturn: &runtimesv1.BidiRunLoadReturn{
					Values: kittehs.TransformMapValues(vs, sdktypes.ToProto),
				},
			},
		}); err != nil {
			return fmt.Errorf("send load return: %w", err)
		}

		return nil

	default:
		return fmt.Errorf("unknown message type: %T", msg)
	}
}

func (c *client) bidiRun1(
	ctx context.Context,
	name sdktypes.Symbol,
	rid sdktypes.RunID,
	path string,
	compiled map[string][]byte,
	globals map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	req := &runtimesv1.BidiRunRequest{
		Request: &runtimesv1.BidiRunRequest_Start1_{
			Start1: &runtimesv1.BidiRunRequest_Start1{
				RuntimeName: name.String(),
				Artifact:    sdktypes.NewBuildArtifact(compiled).ToProto(),
				Data: &runtimesv1.BidiRunRequest_StartData{
					RunId:   rid.String(),
					Path:    path,
					Globals: kittehs.TransformMapValues(globals, sdktypes.ToProto),
				},
			},
		},
	}

	return c.commonBidiRun(ctx, req, rid, cbs)
}

func (c *client) bidiRun(
	ctx context.Context,
	rid sdktypes.RunID,
	path string,
	build *sdkbuildfile.BuildFile,
	globals map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	var buf bytes.Buffer
	if err := build.Write(&buf); err != nil {
		return nil, fmt.Errorf("write build file: %w", err)
	}

	req := &runtimesv1.BidiRunRequest{
		Request: &runtimesv1.BidiRunRequest_Start_{
			Start: &runtimesv1.BidiRunRequest_Start{
				BuildFile: buf.Bytes(),
				Data: &runtimesv1.BidiRunRequest_StartData{
					RunId:   rid.String(),
					Path:    path,
					Globals: kittehs.TransformMapValues(globals, sdktypes.ToProto),
				},
			},
		},
	}

	return c.commonBidiRun(ctx, req, rid, cbs)
}

func (c *client) commonBidiRun(ctx context.Context, req *runtimesv1.BidiRunRequest, rid sdktypes.RunID, cbs *sdkservices.RunCallbacks) (sdkservices.Run, error) {
	// TODO: translate errors from connect to sdkerrors.

	stream := &loggedstream.LoggedStream[runtimesv1.BidiRunResponse, runtimesv1.BidiRunRequest]{
		S:      c.client.BidiRun(ctx),
		SL:     c.sl.With("rid", rid.String()),
		DescRx: bidirunmsgs.DescribeRes,
		DescTx: bidirunmsgs.DescribeReq,
		Level:  zapcore.InfoLevel,
	}

	if err := stream.Send(req); err != nil {
		return nil, fmt.Errorf("send start: %w", err)
	}

	run := &bidiRun{stream: stream, rid: rid, cbs: cbs}

	var startret *runtimesv1.BidiRunLoadReturn

	for {
		msg, err := stream.Receive()
		if err != nil {
			return nil, fmt.Errorf("receive: %w", err)
		}

		if msg.GetStartReturn() != nil {
			startret = msg.GetStartReturn()
			break
		}

		if err := run.handleCallback(ctx, msg); err != nil {
			return nil, fmt.Errorf("callback: %w", err)
		}
	}

	if pb := startret.GetError(); pb != nil {
		perr, err := sdktypes.ProgramErrorFromProto(pb)
		if perr.IsValid() {
			err = perr.ToError()
		} else {
			err = fmt.Errorf("decode error: %w", err)
		}
		return nil, err
	}

	values, err := kittehs.TransformMapValuesError(startret.Values, sdktypes.StrictValueFromProto)
	if err != nil {
		return nil, fmt.Errorf("decode globals: %w", err)
	}

	run.values = values

	return run, nil
}
