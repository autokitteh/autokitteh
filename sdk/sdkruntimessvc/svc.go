package sdkruntimessvc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"connectrpc.com/connect"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	akproto "go.autokitteh.dev/autokitteh/proto"
	runtimesv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runtimes/v1/runtimesv1connect"
	"go.autokitteh.dev/autokitteh/sdk/internal/bidirunmsgs"
	"go.autokitteh.dev/autokitteh/sdk/internal/loggedstream"
	"go.autokitteh.dev/autokitteh/sdk/sdkbuildfile"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdklogger"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type svc struct {
	sl       *zap.SugaredLogger
	runtimes sdkservices.Runtimes

	lsOnce sync.Once
	ls     []sdktypes.Runtime

	runtimesv1connect.UnimplementedRuntimesServiceHandler
}

var _ runtimesv1connect.RuntimesServiceHandler = &svc{}

func Init(sl *zap.SugaredLogger, runtimes sdkservices.Runtimes, mux *http.ServeMux) {
	path, h := runtimesv1connect.NewRuntimesServiceHandler(&svc{runtimes: runtimes, sl: sl})
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
		return nil, sdkerrors.AsConnectError(err)
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
		return nil, sdkerrors.AsConnectError(err)
	}

	rts, err := s.list(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&runtimesv1.ListResponse{Runtimes: kittehs.Transform(rts, sdktypes.ToProto)}), nil
}

func (s *svc) Build(ctx context.Context, req *connect.Request[runtimesv1.BuildRequest]) (*connect.Response[runtimesv1.BuildResponse], error) {
	if err := akproto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	symbols, err := kittehs.TransformError(req.Msg.Symbols, sdktypes.StrictParseSymbol)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	srcFS, err := kittehs.MapToMemFS(req.Msg.Resources)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	bf, err := s.runtimes.Build(ctx, srcFS, symbols, req.Msg.Memo)
	if err != nil {
		if perr, ok := sdktypes.FromError(err); ok {
			return connect.NewResponse(&runtimesv1.BuildResponse{Error: perr.ToProto()}), nil
		}
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	var buf bytes.Buffer
	if err := bf.Write(&buf); err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	return connect.NewResponse(&runtimesv1.BuildResponse{
		BuildFile: buf.Bytes(),
	}), nil
}

func (s *svc) Build1(ctx context.Context, req *connect.Request[runtimesv1.Build1Request]) (*connect.Response[runtimesv1.Build1Response], error) {
	if err := akproto.Validate(req.Msg); err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	symbols, err := kittehs.TransformError(req.Msg.Symbols, sdktypes.StrictParseSymbol)
	if err != nil {
		return nil, sdkerrors.AsConnectError(err)
	}

	srcFS, err := kittehs.MapToMemFS(req.Msg.Resources)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	name, err := sdktypes.Strict(sdktypes.ParseSymbol(req.Msg.RuntimeName))
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	rt, err := s.runtimes.New(ctx, name)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	a, err := rt.Build(ctx, srcFS, req.Msg.Path, symbols)
	if err != nil {
		if perr, ok := sdktypes.FromError(err); ok {
			return connect.NewResponse(&runtimesv1.Build1Response{Error: perr.ToProto()}), nil
		}
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(&runtimesv1.Build1Response{
		Artifact: a.ToProto(),
	}), nil
}

func (s *svc) Run(ctx context.Context, req *connect.Request[runtimesv1.RunRequest], stream *connect.ServerStream[runtimesv1.RunResponse]) error {
	msg := req.Msg

	if err := akproto.Validate(msg); err != nil {
		return sdkerrors.AsConnectError(err)
	}

	rid, err := sdktypes.ParseRunID(msg.RunId)
	if err != nil {
		return sdkerrors.AsConnectError(err)
	}

	gs, err := kittehs.TransformMapValuesError(msg.Globals, sdktypes.StrictValueFromProto)
	if err != nil {
		return sdkerrors.AsConnectError(fmt.Errorf("globals: %w", err))
	}

	bf, err := sdkbuildfile.Read(bytes.NewReader(msg.Artifact))
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("artifact: %w", err))
	}

	cbs := &sdkservices.RunCallbacks{
		Print: func(_ context.Context, _ sdktypes.RunID, msg string) {
			if err := stream.Send(&runtimesv1.RunResponse{Print: msg}); err != nil {
				s.sl.Error("failed to send print message", "err", err)
			}
		},
	}

	run, err := s.runtimes.Run(ctx, rid, msg.Path, bf, gs, cbs)
	if err != nil {
		if perr, ok := sdktypes.FromError(err); ok {
			return stream.Send(&runtimesv1.RunResponse{Error: perr.ToProto()})
		}
		return connect.NewError(connect.CodeUnknown, err)
	}

	return stream.Send(
		&runtimesv1.RunResponse{
			Result: kittehs.TransformMapValues(run.Values(), sdktypes.ToProto),
		},
	)
}

func (s *svc) BidiRun(ctx context.Context, nakedStream *connect.BidiStream[runtimesv1.BidiRunRequest, runtimesv1.BidiRunResponse]) error {
	sl := s.sl

	stream := &loggedstream.LoggedStream[runtimesv1.BidiRunRequest, runtimesv1.BidiRunResponse]{
		SL:     sl,
		S:      nakedStream,
		Level:  zapcore.InfoLevel,
		DescRx: bidirunmsgs.DescribeReq,
		DescTx: bidirunmsgs.DescribeRes,
	}

	sendStartReturnError := func(err error) error {
		if perr, ok := sdktypes.FromError(err); ok {
			return stream.Send(&runtimesv1.BidiRunResponse{
				Response: &runtimesv1.BidiRunResponse_StartReturn{
					StartReturn: &runtimesv1.BidiRunLoadReturn{
						Error: perr.ToProto(),
					},
				},
			})
		}

		return connect.NewError(connect.CodeUnknown, err)
	}

	sendStartReturnResult := func(vs map[string]sdktypes.Value) error {
		return stream.Send(&runtimesv1.BidiRunResponse{
			Response: &runtimesv1.BidiRunResponse_StartReturn{
				StartReturn: &runtimesv1.BidiRunLoadReturn{
					Values: kittehs.TransformMapValues(vs, sdktypes.ToProto),
				},
			},
		})
	}

	sendCallReturnError := func(err error) error {
		if perr, ok := sdktypes.FromError(err); ok {
			return stream.Send(&runtimesv1.BidiRunResponse{
				Response: &runtimesv1.BidiRunResponse_CallReturn{
					CallReturn: &runtimesv1.BidiRunCallReturn{
						Result: &runtimesv1.BidiRunCallReturn_Error{
							Error: perr.ToProto(),
						},
					},
				},
			})
		}

		return connect.NewError(connect.CodeUnknown, err)
	}

	sendCallReturnResult := func(v sdktypes.Value) error {
		return stream.Send(&runtimesv1.BidiRunResponse{
			Response: &runtimesv1.BidiRunResponse_CallReturn{
				CallReturn: &runtimesv1.BidiRunCallReturn{
					Result: &runtimesv1.BidiRunCallReturn_Value{
						Value: v.ToProto(),
					},
				},
			},
		})
	}

	sendPrint := func(msg string) error {
		return stream.Send(&runtimesv1.BidiRunResponse{
			Response: &runtimesv1.BidiRunResponse_Print_{
				Print: &runtimesv1.BidiRunResponse_Print{
					Text: msg,
				},
			},
		})
	}

	sendCall := func(fv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) error {
		return stream.Send(&runtimesv1.BidiRunResponse{
			Response: &runtimesv1.BidiRunResponse_Call{
				Call: &runtimesv1.BidiRunCall{
					Value:  fv.ToProto(),
					Args:   kittehs.Transform(args, sdktypes.ToProto),
					Kwargs: kittehs.TransformMapValues(kwargs, sdktypes.ToProto),
				},
			},
		})
	}

	sendLoad := func(path string) error {
		return stream.Send(&runtimesv1.BidiRunResponse{
			Response: &runtimesv1.BidiRunResponse_Load_{
				Load: &runtimesv1.BidiRunResponse_Load{
					Path: path,
				},
			},
		})
	}

	handleInboundCall := func(run sdkservices.Run, call *runtimesv1.BidiRunCall) error {
		args, err := kittehs.TransformError(call.Args, sdktypes.StrictValueFromProto)
		if err != nil {
			return fmt.Errorf("call args: %w", err)
		}

		kwargs, err := kittehs.TransformMapValuesError(call.Kwargs, sdktypes.StrictValueFromProto)
		if err != nil {
			return fmt.Errorf("call kwargs: %w", err)
		}

		fv, err := sdktypes.StrictValueFromProto(call.Value)
		if err != nil {
			return fmt.Errorf("call value: %w", err)
		}

		v, err := run.Call(ctx, fv, args, kwargs)
		if err != nil {
			return sendCallReturnError(err)
		}

		return sendCallReturnResult(v)
	}

	handleOutboundCall := func(run sdkservices.Run, fv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
		if err := sendCall(fv, args, kwargs); err != nil {
			return sdktypes.InvalidValue, err
		}

		var msg *runtimesv1.BidiRunRequest

		for msg == nil {
			var err error
			if msg, err = stream.Receive(); err != nil {
				return sdktypes.InvalidValue, err
			}

			if call := msg.GetCall(); call != nil {
				// got an inbound call why doing and outbound call.
				if err := handleInboundCall(run, call); err != nil {
					return sdktypes.InvalidValue, err
				}

				msg = nil
			}
		}

		ret := msg.GetCallReturn()
		if ret == nil {
			return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("expected call return message")
		}

		if cerr := ret.GetError(); cerr != nil {
			perr, err := sdktypes.ProgramErrorFromProto(cerr)
			if err != nil {
				err = fmt.Errorf("decode call error: %w", err)
			} else {
				err = perr.ToError()
			}

			return sdktypes.InvalidValue, err
		}

		if v := ret.GetValue(); v != nil {
			return sdktypes.StrictValueFromProto(v)
		}

		return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("invalid call return")
	}

	handleLoad := func(run sdkservices.Run, path string) (map[string]sdktypes.Value, error) {
		if err := sendLoad(path); err != nil {
			return nil, err
		}

		var msg *runtimesv1.BidiRunRequest

		for msg == nil {
			var err error

			if msg, err = stream.Receive(); err != nil {
				return nil, err
			}

			if call := msg.GetCall(); call != nil {
				// got an inbound call while doing a load.
				if err := handleInboundCall(run, call); err != nil {
					return nil, err
				}

				msg = nil
			}
		}

		ret := msg.GetLoadReturn()
		if ret == nil {
			return nil, sdkerrors.NewInvalidArgumentError("expected load return message")
		}

		if cerr := ret.GetError(); cerr != nil {
			perr, err := sdktypes.ProgramErrorFromProto(cerr)
			if err != nil {
				err = fmt.Errorf("decode load error: %w", err)
			} else {
				err = perr.ToError()
			}

			return nil, err
		}

		return kittehs.TransformMapValuesError(ret.GetValues(), sdktypes.StrictValueFromProto)
	}

	handleStart := func() (sdkservices.Run, error) {
		msg, err := stream.Receive()
		if err != nil {
			return nil, fmt.Errorf("receive: %w", err)
		}

		var startData *runtimesv1.BidiRunRequest_StartData

		var start1 *runtimesv1.BidiRunRequest_Start1
		start := msg.GetStart()
		if start != nil {
			startData = start.GetData()
		} else {
			if start1 = msg.GetStart1(); start1 == nil {
				return nil, sdkerrors.NewInvalidArgumentError("expected start message")
			}

			startData = start1.GetData()
		}

		rid, err := sdktypes.Strict(sdktypes.ParseRunID(startData.RunId))
		if err != nil {
			return nil, fmt.Errorf("run_id: %w", err)
		}

		// update logger with run_id so we can track.
		sl = sl.With("run_id", rid)
		stream.SL = sl

		gs, err := kittehs.TransformMapValuesError(startData.Globals, sdktypes.StrictValueFromProto)
		if err != nil {
			return nil, fmt.Errorf("globals: %w", err)
		}

		var run sdkservices.Run

		cbs := &sdkservices.RunCallbacks{
			Print: func(_ context.Context, _ sdktypes.RunID, msg string) {
				if err := sendPrint(msg); err != nil {
					sl.Error("send print", "err", err)
				}
			},
			Call: func(_ context.Context, _ sdktypes.RunID, fv sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
				return handleOutboundCall(run, fv, args, kwargs)
			},
			Load: func(_ context.Context, _ sdktypes.RunID, path string) (map[string]sdktypes.Value, error) {
				return handleLoad(run, path)
			},
			NewRunID: func() sdktypes.RunID {
				// TODO: not sure if even needed.
				sl.DPanic("not implemented")
				return sdktypes.NewRunID() // <-- should be deterministic, which it's not.
			},
		}

		if start != nil {
			bf, err := sdkbuildfile.Read(bytes.NewReader(start.BuildFile))
			if err != nil {
				return nil, fmt.Errorf("artifact: %w", err)
			}

			if run, err = s.runtimes.Run(ctx, rid, startData.Path, bf, gs, cbs); err != nil {
				return nil, sendStartReturnError(err)
			}

		} else if start1 != nil {
			a, err := sdktypes.Strict(sdktypes.BuildArtifactFromProto(start1.Artifact))
			if err != nil {
				return nil, fmt.Errorf("artifact: %w", err)
			}

			name, err := sdktypes.Strict(sdktypes.ParseSymbol(start1.RuntimeName))
			if err != nil {
				return nil, fmt.Errorf("runtime_name: %w", err)
			}

			rt, err := s.runtimes.New(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("new runtime %q: %w", name, err)
			}

			if run, err = rt.Run(ctx, rid, startData.Path, a.CompiledData(), gs, cbs); err != nil {
				return nil, sendStartReturnError(err)
			}
		} else {
			sdklogger.Panic("no start or start1")
		}

		if err := sendStartReturnResult(run.Values()); err != nil {
			return nil, err
		}

		return run, nil
	}

	run, err := handleStart()
	if err != nil {
		if errors.Is(err, io.EOF) {
			// Client disconnected before sending start message.
			sl.Warn("client disconnected before sending start message")
			return nil
		}

		return sdkerrors.AsConnectError(err)
	}

	defer run.Close()

	for {
		sl.Debug("waiting for call")

		msg, err := stream.Receive()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return err
		}

		call := msg.GetCall()
		if call == nil {
			return sdkerrors.NewInvalidArgumentError("expected call message")
		}

		if err := handleInboundCall(run, call); err != nil {
			if errors.Is(err, io.EOF) {
				// Client disconnected before call started.
				sl.Warn("client disconnected before call start")
				return nil
			}

			return sdkerrors.AsConnectError(err)
		}
	}
}
