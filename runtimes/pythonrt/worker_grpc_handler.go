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

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/backend/oauth"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const runnerChResponseTimeout = 5 * time.Second

type workerGRPCHandler struct {
	pb.UnimplementedWorkerServer

	runnerIDsToRuntime map[string]*pySvc
	mu                 *sync.Mutex
}

var w = workerGRPCHandler{
	runnerIDsToRuntime: map[string]*pySvc{},
	mu:                 new(sync.Mutex),
}

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
		return errors.New("unknown runner ID")
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
	defer w.mu.Unlock()

	_, ok := w.runnerIDsToRuntime[req.RunnerId]
	if !ok {
		return &pb.IsActiveRunnerResponse{Error: "unknown runner ID"}, nil
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
		return &pb.LogResponse{Error: "unknown runner ID"}, nil
	}

	m := makeLogMessage(req.Level, req.Message)
	runner.channels.log <- m

	select {
	case <-m.doneChannel:
		return &pb.LogResponse{}, nil
	case <-time.After(runnerChResponseTimeout):
		return &pb.LogResponse{
			Error: "timeout",
		}, nil
	}
}

func (s *workerGRPCHandler) Print(ctx context.Context, req *pb.PrintRequest) (*pb.PrintResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.PrintResponse{Error: "unknown runner ID"}, nil
	}

	m := makeLogMessage(req.Message, "info")

	runner.channels.print <- m
	select {
	case <-m.doneChannel:
		return &pb.PrintResponse{}, nil
	case <-time.After(runnerChResponseTimeout):
		return &pb.PrintResponse{
			Error: "timeout",
		}, nil
	}
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
		return &pb.ActivityResponse{Error: "unknown runner ID"}, nil
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

func makeCallbackMessage(args []sdktypes.Value, kwargs map[string]sdktypes.Value) *callbackMessage {
	callbackChan := make(chan sdktypes.Value, 1)
	errorChannel := make(chan error, 1)

	msg := &callbackMessage{
		args:           args,
		kwargs:         kwargs,
		successChannel: callbackChan,
		errorChannel:   errorChannel,
	}
	return msg
}

func makeLogMessage(message string, level string) *logMessage {
	return &logMessage{
		level:       level,
		message:     message,
		doneChannel: make(chan struct{}, 1),
	}
}

func (s *workerGRPCHandler) Sleep(ctx context.Context, req *pb.SleepRequest) (*pb.SleepResponse, error) {
	if req.DurationMs < 0 {
		return nil, status.Error(codes.InvalidArgument, "negative time")
	}

	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.SleepResponse{Error: "unknown runner ID"}, nil
	}

	secs := float64(req.DurationMs) / 1000.0
	args := []sdktypes.Value{
		sdktypes.NewStringValue("sleep"),
		sdktypes.NewFloatValue(secs),
	}

	msg := makeCallbackMessage(args, nil)
	runner.channels.callback <- msg

	select {
	case err := <-msg.errorChannel:
		err = status.Errorf(codes.Internal, "sleep(%f) -> %s", secs, err)
		return &pb.SleepResponse{Error: err.Error()}, nil
	case <-msg.successChannel:
		return &pb.SleepResponse{}, nil
	}
}

func (s *workerGRPCHandler) StartSession(ctx context.Context, req *pb.StartSessionRequest) (*pb.StartSessionResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.StartSessionResponse{
			Error: "Unknown runner id",
		}, nil
	}

	var data map[string]any
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't unmarshal data: %s", err)
	}

	var memo map[string]string
	if err := json.Unmarshal(req.Memo, &memo); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't unmarshal memo: %s", err)
	}

	vdata, err := kittehs.TransformMapValuesError(data, sdktypes.DefaultValueWrapper.Wrap)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't wrap data: %s", err)
	}

	args := []sdktypes.Value{
		sdktypes.NewStringValue("start"),
		sdktypes.NewStringValue(req.Loc),
		sdktypes.NewDictValueFromStringMap(vdata),
		sdktypes.NewDictValueFromStringMap(kittehs.TransformMapValues(memo, sdktypes.NewStringValue)),
	}
	msg := makeCallbackMessage(args, nil)
	runner.channels.callback <- msg

	select {
	case err := <-msg.errorChannel:
		err = status.Errorf(codes.Internal, "start(%s) -> %s", req.Loc, err)
		return &pb.StartSessionResponse{Error: err.Error()}, nil
	case val := <-msg.successChannel:
		return &pb.StartSessionResponse{SessionId: val.GetString().Value()}, nil
	}
}

func (s *workerGRPCHandler) Subscribe(ctx context.Context, req *pb.SubscribeRequest) (*pb.SubscribeResponse, error) {
	if req.Connection == "" || req.Filter == "" {
		return nil, status.Error(codes.InvalidArgument, "missing connection name or filter")
	}

	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.SubscribeResponse{Error: "unknown runner ID"}, nil
	}

	args := []sdktypes.Value{
		sdktypes.NewStringValue("subscribe"),
		sdktypes.NewStringValue(req.Connection),
		sdktypes.NewStringValue(req.Filter),
	}
	msg := makeCallbackMessage(args, nil)
	runner.channels.callback <- msg

	select {
	case err := <-msg.errorChannel:
		err = status.Errorf(codes.Internal, "subscribe(%s, %s) -> %s", req.Connection, req.Filter, err)
		return &pb.SubscribeResponse{Error: err.Error()}, nil
	case val := <-msg.successChannel:
		signalID := val.GetString().Value()
		return &pb.SubscribeResponse{SignalId: signalID}, nil
	}
}

func (s *workerGRPCHandler) NextEvent(ctx context.Context, req *pb.NextEventRequest) (*pb.NextEventResponse, error) {
	if len(req.SignalIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "at least one signal ID required")
	}
	if req.TimeoutMs < 0 {
		return nil, status.Error(codes.InvalidArgument, "timeout < 0")
	}

	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.NextEventResponse{Error: "unknown runner ID"}, nil
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

	msg := makeCallbackMessage(args, kw)
	runner.channels.callback <- msg

	select {
	case err := <-msg.errorChannel:
		err = status.Errorf(codes.Internal, "next_event(%s, %d) -> %s", req.SignalIds, req.TimeoutMs, err)
		return &pb.NextEventResponse{Error: err.Error()}, nil
	case val := <-msg.successChannel:
		out, err := sdktypes.ValueWrapper{SafeForJSON: true}.Unwrap(val)
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
}

func (s *workerGRPCHandler) Unsubscribe(ctx context.Context, req *pb.UnsubscribeRequest) (*pb.UnsubscribeResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.UnsubscribeResponse{Error: "Unknown runner id"}, nil
	}

	args := []sdktypes.Value{
		sdktypes.NewStringValue("unsubsribe"),
		sdktypes.NewStringValue(req.SignalId),
	}
	msg := makeCallbackMessage(args, nil)
	runner.channels.callback <- msg

	select {
	case err := <-msg.errorChannel:
		err = status.Errorf(codes.Internal, "subscribe(%s) -> %s", req.SignalId, err)
		return &pb.UnsubscribeResponse{Error: err.Error()}, err
	case <-msg.successChannel:
		return &pb.UnsubscribeResponse{}, nil
	}
}

func (s *workerGRPCHandler) RefreshOAuthToken(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		return &pb.RefreshResponse{Error: "unknown runner ID"}, nil
	}

	// Get the integration's OAuth configuration.
	cfg, _, err := oauth.New(runner.log).Get(ctx, req.Integration)
	if err != nil {
		return &pb.RefreshResponse{Error: err.Error()}, nil
	}

	refreshToken, ok := runner.envVars[req.Connection+"__oauth_RefreshToken"]
	if !ok {
		return &pb.RefreshResponse{Error: "missing refresh token"}, nil
	}
	// Get a fresh access token.
	t := &oauth2.Token{RefreshToken: refreshToken}
	t, err = cfg.TokenSource(ctx, t).Token()
	if err != nil {
		return &pb.RefreshResponse{Error: err.Error()}, nil
	}

	return &pb.RefreshResponse{
		Token:   t.AccessToken,
		Expires: timestamppb.New(t.Expiry),
	}, nil
}
