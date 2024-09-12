// gRPC server that accepts calls from the Python runner
package pythonrt

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type remoteSvc struct {
	pb.UnimplementedWorkerServer

	log       *zap.Logger
	cbs       *sdkservices.RunCallbacks
	runID     sdktypes.RunID
	xid       sdktypes.ExecutorID
	syscallFn sdktypes.Value
	lis       net.Listener
	srv       *grpc.Server
	port      int
	runner    pb.RunnerClient

	// We need the right context to send to cbs
	// It can be either the initial context (start of flow) or call context (current activity)
	// We don't have nested activities
	initialCtx context.Context
	callCtx    context.Context

	// One of these will signal end of execution
	done chan *pb.DoneRequest
}

func newRemoteSvc(log *zap.Logger, cbs *sdkservices.RunCallbacks, runID sdktypes.RunID, xid sdktypes.ExecutorID, syscallFn sdktypes.Value) *remoteSvc {
	svc := remoteSvc{
		log:   log,
		cbs:   cbs,
		runID: runID,
		xid:   xid,

		// Buffered so gRPC handler won't get stuck
		done: make(chan *pb.DoneRequest, 1),
	}

	if syscallFn.IsValid() {
		svc.syscallFn = syscallFn
	}

	return &svc
}

func (s *remoteSvc) Health(context.Context, *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{}, nil
}

func pyLevelToZap(level string) zapcore.Level {
	switch level {
	case "DEBUG":
		return zap.DebugLevel
	case "INFO":
		return zap.InfoLevel
	case "WARNING":
		return zap.WarnLevel
	case "ERROR":
		return zap.ErrorLevel
	}

	return zap.InfoLevel
}

func (s *remoteSvc) Log(ctx context.Context, req *pb.LogRequest) (*pb.LogResponse, error) {
	if req.Level == "" {
		return nil, status.Error(codes.InvalidArgument, "empty level")
	}

	level := pyLevelToZap(req.Level)
	s.log.Log(level, req.Message, zap.String("source", "python"))
	return &pb.LogResponse{}, nil
}

func (s *remoteSvc) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
	s.cbs.Print(s.ctx(), s.runID, req.Message)
	return &pb.PrintResponse{}, nil
}

// ak functions

func (s *remoteSvc) Sleep(ctx context.Context, req *pb.SleepRequest) (*pb.SleepResponse, error) {
	if req.DurationMs < 0 {
		return nil, status.Error(codes.InvalidArgument, "negative time")
	}

	secs := float64(req.DurationMs) / 1000.0
	args := []sdktypes.Value{
		sdktypes.NewStringValue("sleep"),
		sdktypes.NewFloatValue(secs),
	}
	_, err := s.cbs.Call(s.ctx(), s.runID, s.syscallFn, args, nil)
	var resp pb.SleepResponse
	if err != nil {
		resp.Error = err.Error()
		err = status.Errorf(codes.Internal, "sleep(%f) -> %s", secs, err)
	}

	return &resp, err
}

func (s *remoteSvc) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	if req.Connection == "" || req.Filter == "" {
		return nil, status.Error(codes.InvalidArgument, "missing connection name or filter")
	}

	args := []sdktypes.Value{
		sdktypes.NewStringValue("subscribe"),
		sdktypes.NewStringValue(req.Connection),
		sdktypes.NewStringValue(req.Filter),
	}
	out, err := s.cbs.Call(s.ctx(), s.runID, s.syscallFn, args, nil)
	if err != nil {
		err = status.Errorf(codes.Internal, "subscribe(%s, %s) -> %s", req.Connection, req.Filter, err)
		return &pb.SubscribeResponse{Error: err.Error()}, err
	}

	signalID := out.GetString().Value()
	resp := pb.SubscribeResponse{SignalId: signalID}
	return &resp, nil
}

func (s *remoteSvc) NextEvent(ctx context.Context, req *pb.NextEventRequest) (*pb.NextEventResponse, error) {
	if len(req.SignalIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one signal ID required")
	}
	if req.TimeoutMs < 0 {
		return nil, status.Error(codes.InvalidArgument, "timeout < 0")
	}

	args := make([]sdktypes.Value, len(req.SignalIds)+1)
	args[0] = sdktypes.NewStringValue("next_event")
	for i, id := range req.SignalIds {
		args[i+1] = sdktypes.NewStringValue(id)
	}
	// timeout is kw only
	kw := make(map[string]sdktypes.Value)
	if req.TimeoutMs > 0 {
		kw["timeout"] = sdktypes.NewFloatValue(float64(req.TimeoutMs) / 1000.0)
	}

	val, err := s.cbs.Call(s.ctx(), s.runID, s.syscallFn, args, kw)
	if err != nil {
		err = status.Errorf(codes.Internal, "next_event(%s, %d) -> %s", req.SignalIds, req.TimeoutMs, err)
		return &pb.NextEventResponse{Error: err.Error()}, err
	}

	out, err := val.Unwrap()
	if err != nil {
		err = status.Errorf(codes.Internal, "can't unwrap %v - %s", val, err)
		return &pb.NextEventResponse{Error: err.Error()}, err
	}

	data, err := json.Marshal(out)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't json.Marshal %v - %s", out, err)
		return &pb.NextEventResponse{Error: err.Error()}, err
	}

	resp := pb.NextEventResponse{
		Event: &pb.Event{
			Data: data,
		},
	}
	return &resp, nil
}

func (s *remoteSvc) Unsubscribe(ctx context.Context, req *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
	args := []sdktypes.Value{
		sdktypes.NewStringValue("unsubsribe"),
		sdktypes.NewStringValue(req.SignalId),
	}
	_, err := s.cbs.Call(s.ctx(), s.runID, s.syscallFn, args, nil)
	if err != nil {
		err = status.Errorf(codes.Internal, "subscribe(%s) -> %s", req.SignalId, err)
		return &pb.UnsubscribeResponse{Error: err.Error()}, err
	}

	return &pb.UnsubscribeResponse{}, nil
}

func (s *remoteSvc) call(val sdktypes.Value) {
	req := pb.ActivityReplyRequest{}

	// We want to send reply in any case
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		reply, err := s.runner.ActivityReply(ctx, &req)
		switch {
		case err != nil:
			s.log.Error("activity reply error", zap.Error(err))
		case reply.Error != "":
			s.log.Error("activity reply error", zap.String("error", reply.Error))
		}
	}()

	if !val.IsFunction() {
		s.log.Error("bad function", zap.Any("val", val))
		req.Error = fmt.Sprintf("%#v is not a function", val)
		return
	}

	fn := val.GetFunction()
	req.Data = fn.Data()
	out, err := s.cbs.Call(s.ctx(), s.runID, val, nil, nil)

	switch {
	case err != nil:
		req.Error = fmt.Sprintf("%s - %s", fn.Name().String(), err)
		s.log.Error("activity reply error", zap.Error(err))
	case !out.IsBytes():
		req.Error = fmt.Sprintf("call output not bytes: %#v", out)
		s.log.Error("activity reply error", zap.String("error", req.Error))
	default:
		data := out.GetBytes().Value()
		req.Result = data
	}
}

// Runner starting activity
func (s *remoteSvc) Activity(ctx context.Context, req *pb.ActivityRequest) (*pb.ActivityResponse, error) {
	fnName := req.CallInfo.Function
	s.log.Info("activity", zap.String("function", fnName))
	fn, err := sdktypes.NewFunctionValue(s.xid, fnName, req.Data, nil, pyModuleFunc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "new function value: %s", err)
	}

	go s.call(fn)

	return &pb.ActivityResponse{}, nil
}

func (s *remoteSvc) Done(ctx context.Context, req *pb.DoneRequest) (*pb.DoneResponse, error) {
	s.done <- req
	return &pb.DoneResponse{}, nil
}

type Healther interface {
	Health(ctx context.Context, in *pb.HealthRequest, opts ...grpc.CallOption) (*pb.HealthResponse, error)
}

func waitForServer(name string, h Healther, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	start := time.Now()
	var req pb.HealthRequest

	for time.Since(start) <= timeout {
		resp, err := h.Health(ctx, &req)
		if err != nil || resp.Error != "" {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		return nil
	}

	return fmt.Errorf("%s not ready after %v", name, timeout)
}

func freePort() (int, error) {
	conn, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}

	conn.Close()
	return conn.Addr().(*net.TCPAddr).Port, nil
}

// Start starts the server on a free port in a new goroutine.
// It returns the port the server listens on.
func (s *remoteSvc) Start() error {
	port, err := freePort()
	if err != nil {
		return err
	}
	s.port = port

	addr := fmt.Sprintf(":%d", port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.lis = lis

	srv := grpc.NewServer(grpc.UnaryInterceptor(newInterceptor(s.log)))
	pb.RegisterWorkerServer(srv, s)
	reflection.Register(srv)

	s.log.Info("server starting", zap.String("address", addr))

	go func() {
		if err := srv.Serve(lis); err != nil {
			s.log.Error("serve gRPC", zap.Error(err))
		}
	}()

	clientAddr := fmt.Sprintf("localhost:%d", s.port)
	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient(clientAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}

	client := pb.NewWorkerClient(conn)
	return waitForServer("worker", client, time.Second)
}

func (s *remoteSvc) Close() error {
	if s.srv == nil {
		return nil
	}

	s.srv.Stop()

	if err := s.lis.Close(); err != nil {
		return fmt.Errorf("close listener - %w", err)
	}

	return nil
}

func (s *remoteSvc) ctx() context.Context {
	if s.callCtx != nil {
		return s.callCtx
	}

	return s.initialCtx
}

func newInterceptor(log *zap.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	fn := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log.Info("call", zap.String("method", info.FullMethod))

		return handler(ctx, req)
	}

	return fn
}
