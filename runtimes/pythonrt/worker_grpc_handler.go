// gRPC server that accepts calls from the Python runner
package pythonrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type workerGRPCHandler struct {
	pb.UnimplementedWorkerServer

	runnerIDsToRuntime map[string]*pySvc
	mu                 *sync.Mutex
}

var w = workerGRPCHandler{
	runnerIDsToRuntime: map[string]*pySvc{},
	mu:                 new(sync.Mutex),
}

// GRPC Server Handling
// func newInterceptor(log *zap.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
// 	fn := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
// 		log.Info("call", zap.String("method", info.FullMethod))

// 		return handler(ctx, req)
// 	}

// 	return fn
// }

func ConfigureWorkerGRPCHandler(l *zap.Logger, mux *http.ServeMux) {
	srv := grpc.NewServer()
	pb.RegisterWorkerServer(srv, &w)
	path := fmt.Sprintf("/%s/", pb.Worker_ServiceDesc.ServiceName)
	mux.Handle(path, srv)
}

func addRunnerToServer(runnerID string, svc *pySvc) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, ok := w.runnerIDsToRuntime[runnerID]
	if ok {
		return errors.New("already registered")
	}
	w.runnerIDsToRuntime[runnerID] = svc
	return nil
}

func removeRunnerFromServer(runnerID string) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	_, ok := w.runnerIDsToRuntime[runnerID]
	if !ok {
		return errors.New("unknown runner id")
	}
	delete(w.runnerIDsToRuntime, runnerID)
	return nil
}

// GRPC Handlers
// TODO: call temporal to verify workflow is still active ?
// TODO: add runner ID to health check so we can verify it
func (s *workerGRPCHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{}, nil
}

func (s *workerGRPCHandler) IsActiveRunner(ctx context.Context, req *pb.IsActiveRunnerRequest) (*pb.IsActiveRunnerResponse, error) {
	w.mu.Lock()
	_, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.IsActiveRunnerResponse{Error: "runner id unknown"}, nil
	}
	return &pb.IsActiveRunnerResponse{}, nil
}

func (s *workerGRPCHandler) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	if req.Level == "" {
		return nil, status.Error(codes.InvalidArgument, "empty level")
	}

	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.LogResponse{
			Error: "Unknown runner id",
		}, nil
	}

	runner.channels.log <- req
	return &pb.LogResponse{}, nil
}

func (s *workerGRPCHandler) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.PrintResponse{
			Error: "Unknown runner id",
		}, nil
	}

	runner.channels.print <- req
	return &pb.PrintResponse{}, nil
}

func (s *workerGRPCHandler) Done(ctx context.Context, req *pb.DoneRequest) (*pb.DoneResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.DoneResponse{}, nil
	}
	runner.channels.done <- req
	return &pb.DoneResponse{}, nil
}

// Runner starting activity
func (s *workerGRPCHandler) Activity(ctx context.Context, req *pb.ActivityRequest) (*pb.ActivityResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.ActivityResponse{
			Error: "runner id unknown",
		}, nil
	}

	fnName := req.CallInfo.Function

	runner.log.Info("activity", zap.String("function", fnName))
	_, err := sdktypes.NewFunctionValue(runner.xid, fnName, req.Data, nil, pyModuleFunc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "new function value: %s", err)
	}

	runner.channels.request <- req

	return &pb.ActivityResponse{}, nil
}

// ak functions

func rpcSyscall[Resp proto.Message](
	rid string,
	callf func(context.Context, sdkservices.RunSyscalls) (sdktypes.Value, error),
	errf func(error) Resp,
	vf func(sdktypes.Value) Resp,
) (Resp, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[rid]
	w.mu.Unlock()
	if !ok {
		return errf(errors.New("runner id unknown")), nil
	}

	callbackChan := make(chan sdktypes.Value)
	errorChannel := make(chan error)

	runner.channels.callback <- &callbackMessage{
		f:              callf,
		successChannel: callbackChan,
		errorChannel:   errorChannel,
	}

	select {
	case err := <-errorChannel:
		return errf(status.Errorf(codes.Internal, "syscall -> %v", err)), nil
	case v := <-callbackChan:
		return vf(v), nil
	}
}

func (s *workerGRPCHandler) Sleep(ctx context.Context, req *pb.SleepRequest) (*pb.SleepResponse, error) {
	if req.DurationMs < 0 {
		return nil, status.Error(codes.InvalidArgument, "negative time")
	}

	return rpcSyscall(
		req.RunnerId,
		func(ctx context.Context, syscalls sdkservices.RunSyscalls) (sdktypes.Value, error) {
			return sdktypes.Nothing, syscalls.Sleep(ctx, time.Duration(req.DurationMs)*time.Millisecond)
		},
		func(err error) *pb.SleepResponse { return &pb.SleepResponse{Error: err.Error()} },
		func(v sdktypes.Value) *pb.SleepResponse { return &pb.SleepResponse{} },
	)
}

func (s *workerGRPCHandler) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	if req.Connection == "" || req.Filter == "" {
		return nil, status.Error(codes.InvalidArgument, "missing connection name or filter")
	}

	return rpcSyscall(
		req.RunnerId,
		func(ctx context.Context, syscalls sdkservices.RunSyscalls) (sdktypes.Value, error) {
			id, err := syscalls.Subscribe(ctx, req.Connection, req.Filter)
			if err != nil {
				return sdktypes.InvalidValue, err
			}

			return sdktypes.NewStringValue(id.String()), nil
		},
		func(err error) *pb.SubscribeResponse { return &pb.SubscribeResponse{Error: err.Error()} },
		func(v sdktypes.Value) *pb.SubscribeResponse {
			return &pb.SubscribeResponse{SignalId: v.GetString().Value()}
		},
	)
}

func (s *workerGRPCHandler) NextEvent(ctx context.Context, req *pb.NextEventRequest) (*pb.NextEventResponse, error) {
	if len(req.SignalIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one signal ID required")
	}
	if req.TimeoutMs < 0 {
		return nil, status.Error(codes.InvalidArgument, "timeout < 0")
	}

	uuids, err := kittehs.TransformError(req.SignalIds, uuid.Parse)
	if err != nil {
		return &pb.NextEventResponse{Error: err.Error()}, nil
	}

	return rpcSyscall(
		req.RunnerId,
		func(ctx context.Context, syscalls sdkservices.RunSyscalls) (sdktypes.Value, error) {
			return syscalls.NextEvent(ctx, uuids, time.Duration(req.TimeoutMs)*time.Millisecond)
		},
		func(err error) *pb.NextEventResponse { return &pb.NextEventResponse{Error: err.Error()} },
		func(val sdktypes.Value) *pb.NextEventResponse {
			out, err := val.Unwrap()
			if err != nil {
				err = status.Errorf(codes.Internal, "can't unwrap %v - %s", val, err)
				return &pb.NextEventResponse{Error: err.Error()}
			}

			data, err := json.Marshal(out)
			if err != nil {
				err = status.Errorf(codes.Internal, "can't json.Marshal %v - %s", out, err)
				return &pb.NextEventResponse{Error: err.Error()}
			}

			return &pb.NextEventResponse{
				Event: &pb.Event{
					Data: data,
				},
			}
		},
	)
}

func (s *workerGRPCHandler) Unsubscribe(ctx context.Context, req *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
	sigid, err := uuid.Parse(req.SignalId)
	if err != nil {
		return &pb.UnsubscribeResponse{
			Error: err.Error(),
		}, nil
	}

	return rpcSyscall(
		req.RunnerId,
		func(ctx context.Context, syscalls sdkservices.RunSyscalls) (sdktypes.Value, error) {
			return sdktypes.Nothing, syscalls.Unsubscribe(ctx, sigid)
		},
		func(err error) *pb.UnsubscribeResponse { return &pb.UnsubscribeResponse{Error: err.Error()} },
		func(sdktypes.Value) *pb.UnsubscribeResponse { return &pb.UnsubscribeResponse{} },
	)
}
