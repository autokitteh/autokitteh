package pythonrt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"path"
	"runtime"
	"slices"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	pbModule "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/module/v1"
	pbUserCode "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/user_code/v1"
	pbValues "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	desc = kittehs.Must1(sdktypes.StrictRuntimeFromProto(&sdktypes.RuntimePB{
		Name:           "python",
		FileExtensions: []string{"py"},
	}))

	venvPath = path.Join(xdg.DataHomeDir(), "venv")
	venvPy   = path.Join(venvPath, "bin", "python")
)

type callbackResponse struct {
	value any
	err   error
}

type callbackMessage struct {
	name string
	fn   func(context.Context, *sdkservices.RunCallbacks, sdktypes.RunID) (any, error)
	ch   chan callbackResponse
}

type logMessage struct {
	message     string
	level       string
	doneChannel chan struct{}
}

type comChannels struct {
	done     chan *pbUserCode.DoneRequest
	err      chan string
	request  chan *pbUserCode.ActivityRequest
	execute  chan *pbUserCode.ExecuteReplyRequest
	print    chan *logMessage
	log      chan *logMessage
	callback chan *callbackMessage
}

type pySvc struct {
	cfg       *Config
	ctx       context.Context
	log       *zap.Logger
	xid       sdktypes.ExecutorID
	runID     sdktypes.RunID
	sessionID sdktypes.SessionID
	cbs       *sdkservices.RunCallbacks
	exports   map[string]sdktypes.Value
	fileName  string // main user code file name (entry point)
	envVars   map[string]string

	runner   *RunnerClient
	runnerID string

	firstCall bool // first call is the trigger, other calls are activities

	channels comChannels

	didCleanup bool
	printDone  chan struct{}
}

func (py *pySvc) cleanup(ctx context.Context) {
	ctx, span := telemetry.T().Start(ctx, "pythonrt.cleanup")
	defer span.End()

	if py.didCleanup {
		return
	}

	py.didCleanup = true

	if err := runnerManager.Stop(ctx, py.runnerID); err != nil {
		py.log.Warn("stop manager", zap.Error(err))
	}

	if err := py.runner.Close(); err != nil {
		py.log.Warn("close runner", zap.Error(err))
	}

	if err := removeRunnerFromServer(py.runnerID); err != nil {
		py.log.Warn("remove runner from grpc", zap.Error(err))
	}
}

func New(
	cfg *Config,
	l *zap.Logger,
	getLocalAddr func() string,
) (*sdkruntimes.Runtime, error) {
	switch cfg.RunnerType {
	case "docker":
		if cfg.WorkerAddress == "" {
			return nil, errors.New("worker address is required for docker runner")
		}
		if err := configureDockerRunnerManager(l, DockerRuntimeConfig{
			LogRunnerCode: cfg.LogRunnerCode,
			LogBuildCode:  cfg.LogBuildCode,
			WorkerAddressProvider: func() string {
				_, port, _ := net.SplitHostPort(getLocalAddr())
				return fmt.Sprintf("%s:%s", cfg.WorkerAddress, port)
			},
		}); err != nil {
			return nil, fmt.Errorf("configure docker runner manager: %w", err)
		}
		l.Info("docker runner configured")
	case "remote":
		if len(cfg.RemoteRunnerEndpoints) == 0 {
			return nil, errors.New("remote runner is enabled but no runner endpoints provided")
		}
		if err := configureRemoteRunnerManager(RemoteRuntimeConfig{
			ManagerAddress: cfg.RemoteRunnerEndpoints,
			WorkerAddress:  cfg.WorkerAddress,
		}); err != nil {
			return nil, fmt.Errorf("configure remote runner manager: %w", err)
		}
		l.Info("remote runner configured")
	default:
		if err := configureLocalRunnerManager(l,
			LocalRunnerManagerConfig{
				WorkerAddress:         cfg.WorkerAddress,
				LazyLoadVEnv:          cfg.LazyLoadLocalVEnv,
				WorkerAddressProvider: getLocalAddr,
				LogCodeRunnerCode:     cfg.LogRunnerCode,
			},
		); err != nil {
			return nil, fmt.Errorf("configure local runner manager: %w", err)
		}

		l.Info("local runner configured")
	}

	return &sdkruntimes.Runtime{
		Desc: desc,
		New:  func() (sdkservices.Runtime, error) { return newSvc(cfg, l) },
	}, nil
}

func newSvc(cfg *Config, l *zap.Logger) (sdkservices.Runtime, error) {
	l = l.With(zap.String("runtime", "python"))

	svc := pySvc{
		cfg:       cfg,
		log:       l,
		firstCall: true,
		channels: comChannels{
			done:     make(chan *pbUserCode.DoneRequest, 1),
			err:      make(chan string, 1),
			request:  make(chan *pbUserCode.ActivityRequest, 1),
			execute:  make(chan *pbUserCode.ExecuteReplyRequest, 1),
			print:    make(chan *logMessage, 1024),
			log:      make(chan *logMessage, 1024),
			callback: make(chan *callbackMessage, 1),
		},
	}

	return &svc, nil
}

func (py *pySvc) Get() sdktypes.Runtime { return desc }

const archiveKey = "code.tar"

// All Python handler function get all event information.
var pyModuleFunc = kittehs.Must1(sdktypes.ModuleFunctionFromProto(&sdktypes.ModuleFunctionPB{
	Input: []*sdktypes.ModuleFunctionFieldPB{
		{Name: "data"},
		{Name: "session_id"},
	},
}))

func entriesToValues(xid sdktypes.ExecutorID, entries []*pbUserCode.Export) (map[string]sdktypes.Value, error) {
	values := make(map[string]sdktypes.Value)
	for _, export := range entries {
		modPB := sdktypes.ModuleFunctionPB{
			Input: make([]*pbModule.FunctionField, len(export.Args)),
		}

		for i, name := range export.Args {
			modPB.Input[i] = &pbModule.FunctionField{
				Name: name,
			}
		}

		modFunc := kittehs.Must1(sdktypes.ModuleFunctionFromProto(&modPB))
		fn, err := sdktypes.NewFunctionValue(xid, export.Name, nil, nil, modFunc)
		if err != nil {
			return nil, err
		}
		values[export.Name] = fn
	}

	return values, nil
}

// entryPointFileName strips the handler from the entry point
// "program.py:on_event" -> "program.py"
func entryPointFileName(entryPoint string) string {
	i := strings.Index(entryPoint, ":")
	if i > 0 {
		return entryPoint[:i]
	}

	return entryPoint
}

func (py *pySvc) handlePrint(ctx context.Context, msg *logMessage) {
	if err := py.cbs.Print(ctx, py.runID, msg.message); err != nil {
		py.log.Error("print error", zap.Error(err))
	}

	close(msg.doneChannel)
}

func (py *pySvc) printConsumer(ctx context.Context) {
	for {
		select {
		case p, ok := <-py.channels.print:
			if !ok {
				py.log.Error("print consumer stopped by closing print channel")
				return
			}

			py.handlePrint(ctx, p)
		case <-py.printDone:
			py.log.Info("print consumer stopped by closing printDone channel")
			return
		}
	}
}

/*
Run starts a Python workflow.

It'll load the Python module and set the list of exported names.
mainPath is in the form `issues.py:on_issue`, Python will load the `issues` module.
Run *does not* execute a function in the Python module, this happens in Call.
*/
func (py *pySvc) Run(
	ctx context.Context,
	runID sdktypes.RunID,
	sessionID sdktypes.SessionID,
	mainPath string,
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	ctx, runSpan := telemetry.T().Start(ctx, "pythonrt.Run")
	defer runSpan.End()

	runnerOK := false
	py.ctx = ctx
	py.runID = runID
	py.sessionID = sessionID
	py.xid = sdktypes.NewExecutorID(runID) // Should be first
	py.log = py.log.With(
		zap.String("run_id", runID.String()),
		zap.String("session_id", sessionID.String()),
		zap.String("path", mainPath),
	)

	py.cbs = cbs

	// Load environment defined by user in the `vars` section of the manifest,
	// these are injected to the Python subprocess environment.
	env, err := cbs.Load(ctx, runID, "env")
	if err != nil {
		return nil, fmt.Errorf("can't load env : %w", err)
	}
	py.envVars = kittehs.TransformMap(env, func(key string, value sdktypes.Value) (string, string) {
		return key, value.GetString().Value()
	})

	tarData := compiled[archiveKey]
	if tarData == nil {
		return nil, fmt.Errorf("%q note found in compiled data", archiveKey)
	}

	startDone := make(chan struct{})

	if t := py.cfg.DelayedStartPrintTimeout; t != 0 {
		go func() {
			select {
			case <-startDone:
				// nop
			case <-time.After(t):
				_ = cbs.Print(ctx, runID, "ᓚᘏᗢ Python runs might take a while to start when running for the first time on a new runner, hang on!")
			}
		}()
	}

	runnerID, runner, err := runnerManager.Start(ctx, sessionID, tarData, py.envVars)
	close(startDone)
	if err != nil {
		return nil, fmt.Errorf("starting runner: %w", err)
	}

	defer func() {
		if !runnerOK {
			py.cleanup(ctx)
		}
	}()

	if err := addRunnerToServer(runnerID, py); err != nil {
		return nil, err
	}

	py.runner = runner
	py.runnerID = runnerID
	defer func() {
		if runnerOK {
			return
		}
		py.cleanup(ctx)
	}()

	py.fileName = entryPointFileName(mainPath)
	req := pbUserCode.ExportsRequest{
		FileName: py.fileName,
	}

	resp, err := py.runner.Exports(ctx, &req)
	if err != nil {
		return nil, err
	}

	py.log.Debug("loaded exports", zap.Any("entries", resp.Exports))
	exports, err := entriesToValues(py.xid, resp.Exports)
	if err != nil {
		py.log.Error("can't create module entries", zap.Error(err))
		return nil, fmt.Errorf("can't create module entries: %w", err)
	}
	py.exports = exports

	runnerOK = true // All is good, don't kill Python subprocess.

	py.printDone = make(chan struct{})
	go py.printConsumer(ctx)

	py.log.Info("run created")
	return py, nil
}

func (py *pySvc) ID() sdktypes.RunID              { return py.runID }
func (py *pySvc) ExecutorID() sdktypes.ExecutorID { return py.xid }

func (py *pySvc) Values() map[string]sdktypes.Value {
	return py.exports
}

func (py *pySvc) Close() {
	py.log.Info("closing (not really)")
	// AK calls Close after `Run`, but we need the Python process running for `Call` as well.
	// We kill the Python process once the initial `Call` is completed.
}

func (py *pySvc) kwToEvent(kwargs map[string]sdktypes.Value) (map[string]any, error) {
	unw := sdktypes.ValueWrapper{IgnoreFunctions: true, SafeForJSON: true}.Unwrap
	return kittehs.TransformMapValuesError(kwargs, unw)
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

func (py *pySvc) call(ctx context.Context, val sdktypes.Value, args []sdktypes.Value, kw map[string]sdktypes.Value) {
	ctx, span := telemetry.T().Start(ctx, "pythonrt.call")
	defer span.End()

	var req pbUserCode.ActivityReplyRequest

	ctx, callSpan := telemetry.T().Start(ctx, "pythonrt.call.cbs.call")

	out, err := py.cbs.Call(ctx, py.runID, val, args, kw)

	callSpan.End()

	switch {
	case err != nil:
		py.log.Info("activity reply error", zap.Error(err))
		req.Error = err.Error()
	case !out.IsCustom():
		py.log.Error("activity reply value not Custom", zap.Any("value", out))
		req.Error = fmt.Sprintf("activity reply not custom (%v)", out)
	default:
		req.Result = out.ToProto()
		req.Result.Custom.ExecutorId = py.xid.String()
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	ctx, replySpan := telemetry.T().Start(ctx, "pythonrt.call.ActivityReply")

	if _, err = py.runner.ActivityReply(ctx, &req); err != nil {
		replySpan.End()

		py.log.Error("activity reply error", zap.Error(err))
		py.sendDone(ctx, err)
	} else {
		replySpan.End()
	}
}

func (py *pySvc) sendDone(ctx context.Context, err error) {
	_, replySpan := telemetry.T().Start(ctx, "pythonrt.sendDone")
	defer replySpan.End()

	req := pbUserCode.DoneRequest{
		RunnerId: py.runnerID,
		Error:    err.Error(),
	}
	py.channels.done <- &req
}

func (py *pySvc) startRequest(ctx context.Context, funcName string, eventData []byte) error {
	ctx, span := telemetry.T().Start(ctx, "pythonrt.startRequest")
	defer span.End()

	req := pbUserCode.StartRequest{
		EntryPoint: fmt.Sprintf("%s:%s", py.fileName, funcName),
		Event: &pbUserCode.Event{
			Data: eventData,
		},
	}
	if _, err := py.runner.Start(ctx, &req); err != nil {
		// TODO: Handle traceback
		return err
	}

	return nil
}

func (py *pySvc) setupHealthcheck(ctx context.Context) chan (error) {
	runnerHealthChan := make(chan error, 1)
	healthFn := func() {
		for {
			select {
			case <-ctx.Done():
				py.log.Debug("health check loop stopped")
				return
			case <-time.After(10 * time.Second):
				healthReq := pbUserCode.RunnerHealthRequest{}

				resp, err := py.runner.Health(ctx, &healthReq)
				if err != nil { // no network/lost packet.load? for sanity check the state locally via IPC/signals
					err = runnerManager.RunnerHealth(ctx, py.runnerID)
				} else if resp.Error != "" {
					err = fmt.Errorf("grpc: %s", resp.Error)
				}
				if err != nil {
					py.log.Error("runner health failed", zap.Error(err))

					// TODO: ENG-1675 - cleanup runner junk

					runnerHealthChan <- err
					return
				}

			}
		}
	}
	py.safelyGo("health", healthFn)
	return runnerHealthChan
}

func (py *pySvc) eventData(kwargs map[string]sdktypes.Value) ([]byte, error) {
	event, err := py.kwToEvent(kwargs)
	if err != nil {
		return nil, fmt.Errorf("can't convert event: %w", err)
	}

	keys := slices.Collect(maps.Keys(event))
	py.log.Info("event", zap.Any("keys", keys))

	return json.Marshal(event)
}

func (py *pySvc) tracebackToLocation(traceback []*pbUserCode.Frame) []sdktypes.CallFrame {
	frames := make([]sdktypes.CallFrame, len(traceback))
	for i, f := range traceback {
		var err error
		frames[i], err = sdktypes.CallFrameFromProto(&sdktypes.CallFramePB{
			Name: f.Name,
			Location: &sdktypes.CodeLocationPB{
				Path: f.Filename,
				Row:  f.Lineno,
				Col:  1,
				Name: f.Name,
			},
		})
		if err != nil {
			py.log.Warn("can't translate traceback frame", zap.Error(err))
		}
	}

	return frames
}

// Call handles a function call from autokitteh.
// First used of Call start a workflow, later invocations are activity calls.
func (py *pySvc) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	ctx, span := telemetry.T().Start(ctx, "pythonrt.Call")
	defer span.End()

	fn := v.GetFunction()
	if !fn.IsValid() {
		py.log.Error("call - invalid function", zap.Any("function", v))
		return sdktypes.InvalidValue, fmt.Errorf("%#v is not a function", v)
	}

	fnName := fn.Name().String()
	py.log.Info("call", zap.String("func", fnName))

	span.SetAttributes(attribute.String("function", fnName), attribute.Bool("first_call", py.firstCall))

	var runnerHealthChan chan error
	if py.firstCall {
		py.firstCall = false

		cancellableCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		runnerHealthChan = py.setupHealthcheck(cancellableCtx)

		eventData, err := py.eventData(kwargs)
		if err != nil {
			return sdktypes.InvalidValue, err
		}

		if err := py.startRequest(ctx, fnName, eventData); err != nil {
			// We won't get print from python here, write the error to the session log.
			if err := py.cbs.Print(ctx, py.runID, err.Error()); err != nil {
				py.log.Error("cbs.Print error", zap.Error(err))
			}
			return sdktypes.InvalidValue, fmt.Errorf("start request: %w", err)
		}

		defer func() {
			py.cleanup(context.Background())

			time.Sleep(10 * time.Millisecond) // Give time to print consumer to finish
			close(py.printDone)
		}()

	} else {
		ctx, span := telemetry.T().Start(ctx, "pythonrt.Call.Execute")

		// If we're here, it's an activity call
		req := pbUserCode.ExecuteRequest{Data: fn.Data()}
		resp, err := py.runner.Execute(ctx, &req)
		switch {
		case err != nil:
			span.End()
			return sdktypes.InvalidValue, err
		case resp.Error != "":
			py.log.Warn("activity error", zap.String("error", resp.Error))
		}

		span.End()
	}

	// Wait for client Done or ActivityReplyRequest message
	// This *can't* run in an different goroutine since callbacks to temporal need to be in the same goroutine.
	for {
		ctx, selSpan := telemetry.T().Start(ctx, "pythonrt.Call.select")

		select {
		case r := <-py.channels.log:
			selSpan.End()
			span.AddEvent("log")

			py.log.Log(pyLevelToZap(r.level), r.message)
			close(r.doneChannel)
		case v := <-py.channels.execute:
			selSpan.End()
			span.AddEvent("execute")

			py.log.Info("execute")

			if v.Result == nil {
				py.log.Error("execute: nil result")
				return sdktypes.InvalidValue, errors.New("execute result is nil")
			}

			v.Result.Custom.ExecutorId = py.xid.String()
			return sdktypes.ValueFromProto(v.Result)
		case r := <-py.channels.request:
			selSpan.End()
			span.AddEvent("request")

			var (
				fnName = "pyFunc"
				args   []sdktypes.Value
				kw     map[string]sdktypes.Value
			)

			if r.CallInfo != nil {
				fnName = r.CallInfo.Function
				args = kittehs.Transform(r.CallInfo.Args, func(v *pbValues.Value) sdktypes.Value {
					// TODO(ENG-1838): What if there's an error?
					val, _ := sdktypes.ValueFromProto(v)
					return val
				})
				kw = kittehs.TransformMap(r.CallInfo.Kwargs, func(k string, v *pbValues.Value) (string, sdktypes.Value) {
					// TODO(ENG-1838): What if there's an error?
					val, _ := sdktypes.ValueFromProto(v)
					return k, val
				})
			}

			// it was already checked before we got here
			fn, err := sdktypes.NewFunctionValue(py.xid, fnName, r.Data, nil, pyModuleFunc)
			if err != nil {
				return sdktypes.InvalidValue, err
			}

			py.call(ctx, fn, args, kw)

		case cb := <-py.channels.callback:
			selSpan.End()
			span.AddEvent("callback")

			py.log.Info("syscall", zap.String("name", cb.name))

			ctx, fnSpan := telemetry.T().Start(ctx, "pythonrt.Call.syscall")
			span.SetAttributes(attribute.String("name", cb.name))
			val, err := cb.fn(ctx, py.cbs, py.runID)
			fnSpan.End()

			_, sendSpan := telemetry.T().Start(ctx, "pythonrt.Call.callbackResponse")
			cb.ch <- callbackResponse{value: val, err: err}
			sendSpan.End()

		case healthErr := <-runnerHealthChan:
			selSpan.End()
			span.AddEvent("health")

			if healthErr != nil {
				return sdktypes.InvalidValue, sdkerrors.NewRetryableErrorf("runner health: %w", healthErr)
			}
		case done := <-py.channels.done:
			selSpan.End()
			span.AddEvent("done")

			py.log.Info("done signal")

			if done.Error != "" {
				py.log.Info("done error", zap.String("error", done.Error))
				perr := sdktypes.NewProgramError(
					sdktypes.NewStringValue(done.Error),
					py.tracebackToLocation(done.Traceback),
					map[string]string{"raw": done.Error},
				)
				return sdktypes.InvalidValue, perr.ToError()
			}

			if done.Result == nil {
				py.log.Error("done: nil result")
				return sdktypes.InvalidValue, errors.New("done result is nil")
			}

			done.Result.Custom.ExecutorId = py.xid.String()
			return sdktypes.ValueFromProto(done.Result)
		case <-ctx.Done():
			selSpan.End()

			span.AddEvent("ctx_done")

			return sdktypes.InvalidValue, fmt.Errorf("context expired - %w", ctx.Err())
		}
	}
}

// safelyGo spins a goroutine and guards against panic in it.
// TODO: Should this be in internal/kittehs?
func (py *pySvc) safelyGo(name string, fn func()) {
	go func() {
		defer func() {
			cs := callStack()
			if err := recover(); err != nil {
				py.log.Error(
					"unhandled panic",
					zap.String("name", name),
					zap.Any("error", err),
					zap.Any("stack", cs),
				)

				err := fmt.Errorf("%s: %s\n%s", name, err, cs)
				py.sendDone(context.Background(), err)
			}
		}()

		fn()
	}()
}

// callStack returns the current call stack as string.
func callStack() string {
	pcs := make([]uintptr, 256)
	n := runtime.Callers(2, pcs)
	pcs = pcs[:n]

	var buf bytes.Buffer
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&buf, "%s:%d %s\n", frame.File, frame.Line, frame.Function)

		if !more {
			break
		}
	}

	return buf.String()
}
