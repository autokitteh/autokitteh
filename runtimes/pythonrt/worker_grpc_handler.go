// gRPC server that accepts calls from the Python runner
package pythonrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"go.autokitteh.dev/autokitteh/internal/backend/oauth"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	userCode "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/user_code/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const runnerChResponseTimeout = 5 * time.Second

type workerGRPCHandler struct {
	userCode.HandlerServiceServer

	runnerIDsToRuntime map[string]*pySvc
	mu                 *sync.Mutex
	log                *zap.Logger
}

var w = workerGRPCHandler{
	runnerIDsToRuntime: map[string]*pySvc{},
	mu:                 new(sync.Mutex),
}

func ConfigureWorkerGRPCHandler(l *zap.Logger, mux *http.ServeMux) {
	w.log = l
	srv := grpc.NewServer()
	userCode.RegisterHandlerServiceServer(srv, &w)
	path := fmt.Sprintf("/%s/", userCode.HandlerService_ServiceDesc.ServiceName)
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
func (s *workerGRPCHandler) Health(ctx context.Context, req *userCode.HandlerHealthRequest) (*userCode.HandlerHealthResponse, error) {
	return &userCode.HandlerHealthResponse{}, nil
}

func (s *workerGRPCHandler) IsActiveRunner(ctx context.Context, req *userCode.IsActiveRunnerRequest) (*userCode.IsActiveRunnerResponse, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	_, ok := w.runnerIDsToRuntime[req.RunnerId]
	if !ok {
		w.log.Error("unknown runner ID", zap.String("id", req.RunnerId))
		return &userCode.IsActiveRunnerResponse{Error: "unknown runner ID"}, nil
	}

	return &userCode.IsActiveRunnerResponse{}, nil
}

func (s *workerGRPCHandler) Log(ctx context.Context, req *userCode.LogRequest) (*userCode.LogResponse, error) {
	if req.Level == "" {
		w.log.Error("empty log level")
		return nil, status.Error(codes.InvalidArgument, "empty level")
	}

	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		w.log.Error("unknown runner ID", zap.String("id", req.RunnerId))
		return &userCode.LogResponse{Error: "unknown runner ID"}, nil
	}

	m := &logMessage{level: req.Level, message: req.Message, doneChannel: make(chan struct{})}

	runner.channels.log <- m

	select {
	case <-m.doneChannel:
		return &userCode.LogResponse{}, nil
	case <-time.After(runnerChResponseTimeout):
		return &userCode.LogResponse{
			Error: "timeout",
		}, nil
	}
}

func (s *workerGRPCHandler) Print(ctx context.Context, req *userCode.PrintRequest) (*userCode.PrintResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		w.log.Error("unknown runner ID", zap.String("id", req.RunnerId))
		return &userCode.PrintResponse{Error: "unknown runner ID"}, nil
	}

	m := &logMessage{level: "info", message: req.Message, doneChannel: make(chan struct{})}

	runner.channels.print <- m

	select {
	case <-m.doneChannel:
		return &userCode.PrintResponse{}, nil
	case <-time.After(runnerChResponseTimeout):
		return &userCode.PrintResponse{
			Error: "timeout",
		}, nil
	}
}

func (s *workerGRPCHandler) Done(ctx context.Context, req *userCode.DoneRequest) (*userCode.DoneResponse, error) {
	s.log.Info("Done request", zap.String("error", req.Error))
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		w.log.Error("unknown runner ID", zap.String("id", req.RunnerId))
		return &userCode.DoneResponse{}, nil
	}

	runner.channels.done <- req
	close(runner.channels.done)
	return &userCode.DoneResponse{}, nil
}

// Runner starting activity
func (s *workerGRPCHandler) Activity(ctx context.Context, req *userCode.ActivityRequest) (*userCode.ActivityResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		w.log.Error("unknown runner ID", zap.String("id", req.RunnerId))
		return &userCode.ActivityResponse{Error: "unknown runner ID"}, nil
	}

	fnName := req.CallInfo.Function

	runner.log.Info("activity", zap.String("function", fnName))
	_, err := sdktypes.NewFunctionValue(runner.xid, fnName, req.Data, nil, pyModuleFunc)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "new function value: %s", err)
	}

	runner.channels.request <- req
	return &userCode.ActivityResponse{}, nil
}

// ak functions

func (w *workerGRPCHandler) callback(ctx context.Context, rid string, name string, fn func(context.Context, *sdkservices.RunCallbacks, sdktypes.RunID) (any, error)) (*callbackResponse, error) {
	l := w.log.With(zap.String("runner", rid), zap.String("name", name))

	startedAt := time.Now()

	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[rid]
	w.mu.Unlock()
	if !ok {
		l.Error("unknown runner ID", zap.String("id", rid))
		return nil, errors.New("unknown runner ID")
	}

	l.Debug("sending callback")

	msg := &callbackMessage{
		name: name,
		fn:   fn,
		ch:   make(chan callbackResponse, 1),
	}

	runner.channels.callback <- msg

	l.Debug("callback sent", zap.Duration("duration", time.Since(startedAt)))

	select {
	case resp := <-msg.ch:
		l.Debug("callback response", zap.Any("response", resp), zap.Duration("duration", time.Since(startedAt)))
		return &resp, nil
	case <-ctx.Done():
		l.Debug("context cancelled", zap.Duration("duration", time.Since(startedAt)))
		return nil, ctx.Err()
	}
}

func (s *workerGRPCHandler) Sleep(ctx context.Context, req *userCode.SleepRequest) (*userCode.SleepResponse, error) {
	if req.DurationMs < 0 {
		return nil, status.Error(codes.InvalidArgument, "negative time")
	}
	fn := func(ctx context.Context, cbs *sdkservices.RunCallbacks, rid sdktypes.RunID) (any, error) {
		return nil, cbs.Sleep(ctx, rid, time.Duration(req.DurationMs)*time.Millisecond)
	}

	resp, err := s.callback(ctx, req.RunnerId, "sleep", fn)
	if err != nil {
		return &userCode.SleepResponse{Error: err.Error()}, nil
	}

	if resp.err != nil {
		err = status.Errorf(codes.Internal, "sleep(%v) -> %v", req.DurationMs, err)
		return &userCode.SleepResponse{Error: err.Error()}, nil
	}

	return &userCode.SleepResponse{}, nil
}

func (s *workerGRPCHandler) StartSession(ctx context.Context, req *userCode.StartSessionRequest) (*userCode.StartSessionResponse, error) {
	var data map[string]any
	if err := json.Unmarshal(req.Data, &data); err != nil {
		s.log.Error("unmarshal Data", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "can't unmarshal data: %s", err)
	}

	var memo map[string]string
	if err := json.Unmarshal(req.Memo, &memo); err != nil {
		s.log.Error("marshal Memo", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "can't unmarshal memo: %s", err)
	}

	vdata, err := kittehs.TransformMapValuesError(data, sdktypes.DefaultValueWrapper.Wrap)
	if err != nil {
		s.log.Error("wrapping values", zap.Error(err))
		return nil, status.Errorf(codes.InvalidArgument, "can't wrap data: %s", err)
	}

	loc, err := sdktypes.ParseCodeLocation(req.Loc)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "can't parse code location: %v", err)
	}

	fn := func(ctx context.Context, cbs *sdkservices.RunCallbacks, rid sdktypes.RunID) (any, error) {
		return cbs.Start(ctx, rid, loc, vdata, memo)
	}

	resp, err := s.callback(ctx, req.RunnerId, "start", fn)
	if err != nil {
		return &userCode.StartSessionResponse{Error: err.Error()}, nil
	}

	if resp.err != nil {
		err = status.Errorf(codes.Internal, "start(%s) -> %s", req.Loc, err)
		return &userCode.StartSessionResponse{Error: err.Error()}, nil
	}

	return &userCode.StartSessionResponse{SessionId: resp.value.(sdktypes.SessionID).String()}, nil
}

func (s *workerGRPCHandler) Subscribe(ctx context.Context, req *userCode.SubscribeRequest) (*userCode.SubscribeResponse, error) {
	if req.Connection == "" {
		return nil, status.Error(codes.InvalidArgument, "missing connection name or filter")
	}

	fn := func(ctx context.Context, cbs *sdkservices.RunCallbacks, rid sdktypes.RunID) (any, error) {
		return cbs.Subscribe(ctx, rid, req.Connection, req.Filter)
	}

	resp, err := s.callback(ctx, req.RunnerId, "subscribe", fn)
	if err != nil {
		return &userCode.SubscribeResponse{Error: err.Error()}, nil
	}

	if resp.err != nil {
		err = status.Errorf(codes.Internal, "subscribe(%s, %s) -> %s", req.Connection, req.Filter, err)
		return &userCode.SubscribeResponse{Error: err.Error()}, nil
	}

	return &userCode.SubscribeResponse{SignalId: resp.value.(string)}, nil
}

func (s *workerGRPCHandler) NextEvent(ctx context.Context, req *userCode.NextEventRequest) (*userCode.NextEventResponse, error) {
	if len(req.SignalIds) == 0 {
		return &userCode.NextEventResponse{
			Event: &userCode.Event{
				Data: []byte("null"),
			},
		}, nil
	}

	if req.TimeoutMs < 0 {
		w.log.Error("bad timeout", zap.Int64("timeout", req.TimeoutMs))
		return nil, status.Error(codes.InvalidArgument, "timeout < 0")
	}

	fn := func(ctx context.Context, cbs *sdkservices.RunCallbacks, rid sdktypes.RunID) (any, error) {
		return cbs.NextEvent(ctx, rid, req.SignalIds, time.Duration(req.TimeoutMs)*time.Millisecond)
	}

	resp, err := s.callback(ctx, req.RunnerId, "next_event", fn)
	if err != nil {
		return &userCode.NextEventResponse{Error: err.Error()}, nil
	}

	if resp.err != nil {
		err = status.Errorf(codes.Internal, "next_event(%s, %d) -> %s", req.SignalIds, req.TimeoutMs, err)
		return &userCode.NextEventResponse{Error: err.Error()}, nil
	}

	out, err := sdktypes.ValueWrapper{SafeForJSON: true}.Unwrap(resp.value.(sdktypes.Value))
	if err != nil {
		err = status.Errorf(codes.Internal, "can't unwrap %v - %s", resp.value, err)
		return &userCode.NextEventResponse{Error: err.Error()}, err
	}

	data, err := json.Marshal(out)
	if err != nil {
		err = status.Errorf(codes.Internal, "can't json.Marshal %v - %s", out, err)
		return &userCode.NextEventResponse{Error: err.Error()}, err
	}

	return &userCode.NextEventResponse{
		Event: &userCode.Event{
			Data: data,
		},
	}, nil
}

func (s *workerGRPCHandler) Unsubscribe(ctx context.Context, req *userCode.UnsubscribeRequest) (*userCode.UnsubscribeResponse, error) {
	fn := func(ctx context.Context, cbs *sdkservices.RunCallbacks, rid sdktypes.RunID) (any, error) {
		return nil, cbs.Unsubscribe(ctx, rid, req.SignalId)
	}

	resp, err := s.callback(ctx, req.RunnerId, "unsubscribe", fn)
	if err != nil {
		return &userCode.UnsubscribeResponse{Error: err.Error()}, nil
	}

	if resp.err != nil {
		err = status.Errorf(codes.Internal, "unsubscribe(%s) -> %s", req.SignalId, err)
		return &userCode.UnsubscribeResponse{Error: err.Error()}, nil
	}

	return &userCode.UnsubscribeResponse{}, nil
}

func (s *workerGRPCHandler) EncodeJWT(ctx context.Context, req *userCode.EncodeJWTRequest) (*userCode.EncodeJWTResponse, error) {
	// GitHub's JWTs must be signed using the RS256 algorithm:
	// https://docs.github.com/en/apps/creating-github-apps/authenticating-with-a-github-app/generating-a-json-web-token-jwt-for-a-github-app
	if req.Algorithm != jwt.SigningMethodRS256.Name {
		return &userCode.EncodeJWTResponse{Error: "unsupported signing method: " + req.Algorithm}, nil
	}

	claims := jwt.MapClaims{}
	for key, value := range req.Payload {
		claims[key] = value
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	pem, ok := os.LookupEnv("GITHUB_PRIVATE_KEY")
	if !ok {
		return &userCode.EncodeJWTResponse{Error: "missing GitHub private key"}, nil
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(pem))
	if err != nil {
		return &userCode.EncodeJWTResponse{Error: fmt.Sprintf("invalid GitHub private key: %v", err)}, nil
	}

	signed, err := token.SignedString(key)
	if err != nil {
		return &userCode.EncodeJWTResponse{Error: fmt.Sprintf("failed to sign JWT: %v", err)}, nil
	}
	return &userCode.EncodeJWTResponse{Jwt: signed}, nil
}

func (s *workerGRPCHandler) RefreshOAuthToken(ctx context.Context, req *userCode.RefreshRequest) (*userCode.RefreshResponse, error) {
	w.mu.Lock()
	runner, ok := w.runnerIDsToRuntime[req.RunnerId]
	w.mu.Unlock()
	if !ok {
		w.log.Error("unknown runner ID", zap.String("id", req.RunnerId))
		return &userCode.RefreshResponse{Error: "unknown runner ID"}, nil
	}

	// Get the integration's OAuth configuration.
	// TODO(INT-46): pass vars service instead of nil
	cfg, _, err := oauth.New(runner.log, nil).Get(ctx, req.Integration)
	if err != nil {
		return &userCode.RefreshResponse{Error: err.Error()}, nil
	}

	// Get a fresh access token.
	refreshToken, ok := runner.envVars[req.Connection+"__oauth_RefreshToken"]
	if !ok {
		// New connection variable name, only in Microsoft integrations for now.
		refreshToken, ok = runner.envVars[req.Connection+"__oauth_refresh_token"]
	}
	if !ok {
		return &userCode.RefreshResponse{Error: "missing refresh token"}, nil
	}

	t := &oauth2.Token{RefreshToken: refreshToken}
	t, err = cfg.TokenSource(ctx, t).Token()
	if err != nil {
		return &userCode.RefreshResponse{Error: err.Error()}, nil
	}

	return &userCode.RefreshResponse{
		Token:   t.AccessToken,
		Expires: timestamppb.New(t.Expiry),
	}, nil
}

func (s *workerGRPCHandler) Signal(ctx context.Context, req *userCode.SignalRequest) (*userCode.SignalResponse, error) {
	pbsig := req.Signal

	sid, err := sdktypes.ParseSessionID(req.SessionId)
	if err != nil {
		return &userCode.SignalResponse{Error: fmt.Sprintf("invalid session id: %v", err)}, nil
	}

	payload, err := sdktypes.ValueFromProto(pbsig.Payload)
	if err != nil {
		return &userCode.SignalResponse{Error: fmt.Sprintf("invalid payload: %v", err)}, nil
	}

	fn := func(ctx context.Context, cbs *sdkservices.RunCallbacks, rid sdktypes.RunID) (any, error) {
		return nil, cbs.Signal(ctx, rid, sid, pbsig.Name, payload)
	}

	resp, err := s.callback(ctx, req.RunnerId, "signal", fn)
	if err != nil {
		return &userCode.SignalResponse{Error: err.Error()}, nil
	}

	if resp.err != nil {
		err = status.Errorf(codes.Internal, "signal(%s,%s) -> %s", pbsig.Name, sid.String(), err)
		return &userCode.SignalResponse{Error: err.Error()}, nil
	}

	return &userCode.SignalResponse{}, nil
}

func (s *workerGRPCHandler) NextSignal(ctx context.Context, req *userCode.NextSignalRequest) (*userCode.NextSignalResponse, error) {
	if len(req.Names) == 0 {
		return &userCode.NextSignalResponse{}, nil
	}

	if req.TimeoutMs < 0 {
		w.log.Error("bad timeout", zap.Int64("timeout", req.TimeoutMs))
		return nil, status.Error(codes.InvalidArgument, "timeout < 0")
	}

	fn := func(ctx context.Context, cbs *sdkservices.RunCallbacks, rid sdktypes.RunID) (any, error) {
		return cbs.NextSignal(ctx, rid, req.Names, time.Duration(req.TimeoutMs)*time.Millisecond)
	}

	resp, err := s.callback(ctx, req.RunnerId, "next_signal", fn)
	if err != nil {
		return &userCode.NextSignalResponse{Error: err.Error()}, nil
	}

	if resp.err != nil {
		err = status.Errorf(codes.Internal, "next_signal(%v, %d) -> %s", req.Names, req.TimeoutMs, err)
		return &userCode.NextSignalResponse{Error: err.Error()}, nil
	}

	sig := resp.value.(*sdkservices.RunSignal)
	if sig == nil {
		return &userCode.NextSignalResponse{}, nil
	}

	return &userCode.NextSignalResponse{
		Signal: &userCode.Signal{
			Name:    sig.Name,
			Payload: sig.Payload.ToProto(),
		},
	}, nil
}

func (s *workerGRPCHandler) StoreList(context.Context, *userCode.StoreListRequest) (*userCode.StoreListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "store list")
}

func (s *workerGRPCHandler) StoreGet(context.Context, *userCode.StoreGetRequest) (*userCode.StoreGetResponse, error) {
	return nil, status.Error(codes.Unimplemented, "store get")
}

func (s *workerGRPCHandler) StoreMutate(context.Context, *userCode.StoreMutateRequest) (*userCode.StoreMutateResponse, error) {
	return nil, status.Error(codes.Unimplemented, "store mutate")
}
