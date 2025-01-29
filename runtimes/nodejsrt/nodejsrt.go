package nodejsrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"path"
	"slices"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

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

type callbackMessage struct {
	args           []sdktypes.Value
	kwargs         map[string]sdktypes.Value
	successChannel chan sdktypes.Value
	errorChannel   chan error
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
	print    chan *logMessage
	log      chan *logMessage
	callback chan *callbackMessage
}

type nodejsSvc struct {
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
	// remote       *workerGRPCHandler

	// runner       Runner
	runner *RunnerClient
	// runnerManager pb.RunnerManagerClient
	runnerID string

	firstCall bool // first call is the trigger, other calls are activities

	channels comChannels

	syscallFn sdktypes.Value
}

func (js *nodejsSvc) cleanup(ctx context.Context) {
	if err := runnerManager.Stop(ctx, js.runnerID); err != nil {
		js.log.Warn("stop manager", zap.Error(err))
	}

	if err := js.runner.Close(); err != nil {
		js.log.Warn("close runner", zap.Error(err))
	}

	if err := removeRunnerFromServer(js.runnerID); err != nil {
		js.log.Warn("remove runner from grpc", zap.Error(err))
	}
}

func New(cfg *Config, l *zap.Logger, getLocalAddr func() string) (*sdkruntimes.Runtime, error) {
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

	svc := nodejsSvc{
		cfg:       cfg,
		log:       l,
		firstCall: true,
		channels: comChannels{
			done:     make(chan *pbUserCode.DoneRequest, 1),
			err:      make(chan string, 1),
			request:  make(chan *pbUserCode.ActivityRequest, 1),
			print:    make(chan *logMessage, 1),
			log:      make(chan *logMessage, 1),
			callback: make(chan *callbackMessage, 1),
		},
	}

	return &svc, nil
}

func (js *nodejsSvc) Get() sdktypes.Runtime { return desc }

const archiveKey = "code.tar"

// All Python handler function get all event information.
var pyModuleFunc = kittehs.Must1(sdktypes.ModuleFunctionFromProto(&sdktypes.ModuleFunctionPB{
	Input: []*sdktypes.ModuleFunctionFieldPB{
		{Name: "created_at"},
		{Name: "data"},
		{Name: "event_id"},
		{Name: "integration_id"},
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

func loadSyscall(values map[string]sdktypes.Value) (sdktypes.Value, error) {
	ak, ok := values["ak"]
	if !ok {
		// py.log.Warn("can't find `ak` in values")
		return sdktypes.InvalidValue, nil
	}

	if !ak.IsStruct() {
		return sdktypes.InvalidValue, errors.New("`ak` is not a struct")
	}

	syscall, ok := ak.GetStruct().Fields()["syscall"]
	if !ok {
		return sdktypes.InvalidValue, errors.New("`syscall` not found in `ak`")
	}
	if !syscall.IsFunction() {
		return sdktypes.InvalidValue, errors.New("`syscall` is not a function")
	}

	return syscall, nil
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

/*
Run starts a Python workflow.

It'll load the Python module and set the list of exported names.
mainPath is in the form `issues.py:on_issue`, Python will load the `issues` module.
Run *does not* execute a function in the Python module, this happens in Call.
*/
func (js *nodejsSvc) Run(
	ctx context.Context,
	runID sdktypes.RunID,
	sessionID sdktypes.SessionID,
	mainPath string,
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	runnerOK := false
	js.ctx = ctx
	js.runID = runID
	js.sessionID = sessionID
	js.xid = sdktypes.NewExecutorID(runID) // Should be first
	js.log = js.log.With(
		zap.String("run_id", runID.String()),
		zap.String("session_id", sessionID.String()),
		zap.String("path", mainPath),
	)

	js.cbs = cbs

	// Load environment defined by user in the `vars` section of the manifest,
	// these are injected to the Python subprocess environment.
	env, err := cbs.Load(ctx, runID, "env")
	if err != nil {
		return nil, fmt.Errorf("can't load env : %w", err)
	}
	js.envVars = kittehs.TransformMap(env, func(key string, value sdktypes.Value) (string, string) {
		return key, value.GetString().Value()
	})

	tarData := compiled[archiveKey]
	if tarData == nil {
		return nil, fmt.Errorf("%q note found in compiled data", archiveKey)
	}

	js.syscallFn, err = loadSyscall(values)
	if err != nil {
		return nil, fmt.Errorf("can't load syscall: %w", err)
	}

	runnerID, runner, err := runnerManager.Start(ctx, sessionID, tarData, js.envVars)
	if err != nil {
		return nil, fmt.Errorf("starting runner: %w", err)
	}

	defer func() {
		if !runnerOK {
			js.cleanup(ctx)
		}
	}()

	if err := addRunnerToServer(runnerID, js); err != nil {
		return nil, err
	}

	js.runner = runner
	js.runnerID = runnerID
	defer func() {
		if runnerOK {
			return
		}
		js.cleanup(ctx)
	}()

	js.fileName = entryPointFileName(mainPath)
	req := pbUserCode.ExportsRequest{
		FileName: js.fileName,
	}

	resp, err := js.runner.Exports(ctx, &req)
	if err != nil {
		return nil, err
	}

	js.log.Debug("loaded exports", zap.Any("entries", resp.Exports))
	exports, err := entriesToValues(js.xid, resp.Exports)
	if err != nil {
		js.log.Error("can't create module entries", zap.Error(err))
		return nil, fmt.Errorf("can't create module entries: %w", err)
	}
	js.exports = exports

	runnerOK = true // All is good, don't kill Python subprocess.

	js.log.Info("run created")
	return js, nil
}

func (js *nodejsSvc) ID() sdktypes.RunID              { return js.runID }
func (js *nodejsSvc) ExecutorID() sdktypes.ExecutorID { return js.xid }

func (js *nodejsSvc) Values() map[string]sdktypes.Value {
	return js.exports
}

func (js *nodejsSvc) Close() {
	js.log.Info("closing (not really)")
	// AK calls Close after `Run`, but we need the Python process running for `Call` as well.
	// We kill the Python process once the initial `Call` is completed.
}

func (js *nodejsSvc) kwToEvent(kwargs map[string]sdktypes.Value) (map[string]any, error) {
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

func (js *nodejsSvc) call(ctx context.Context, val sdktypes.Value, args []sdktypes.Value, kw map[string]sdktypes.Value) {
	var req pbUserCode.ActivityReplyRequest

	out, err := js.cbs.Call(js.ctx, js.runID, val, args, kw)
	switch {
	case err != nil:
		js.log.Info("activity reply error", zap.Error(err))
		req.Error = err.Error()
	case !out.IsCustom():
		js.log.Error("activity reply value not Custom", zap.Any("value", out))
		req.Error = fmt.Sprintf("activity reply not custom (%v)", out)
	default:
		req.Result = out.ToProto()
		req.Result.Custom.ExecutorId = js.xid.String()
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	if _, err = js.runner.ActivityReply(ctx, &req); err != nil {
		js.log.Error("activity reply error", zap.Error(err))
		req := pbUserCode.DoneRequest{
			RunnerId: js.runnerID,
			Error:    err.Error(),
		}
		js.channels.done <- &req
	}
}

func (js *nodejsSvc) setupCallbacksListeningLoop(ctx context.Context) chan (error) {
	callbackErrChan := make(chan error, 1)
	go func() {
		for {
			select {
			case r := <-js.channels.log:
				js.log.Log(pyLevelToZap(r.level), r.message)
				close(r.doneChannel)
			case p := <-js.channels.print:
				js.cbs.Print(ctx, js.runID, p.message)
				close(p.doneChannel)
			case r := <-js.channels.request:
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
				fn, err := sdktypes.NewFunctionValue(js.xid, fnName, r.Data, nil, pyModuleFunc)
				if err != nil {
					callbackErrChan <- err
					return
				}
				js.call(ctx, fn, args, kw)
			case cb := <-js.channels.callback:
				val, err := js.cbs.Call(ctx, js.runID, js.syscallFn, cb.args, cb.kwargs)
				if err != nil {
					cb.errorChannel <- err
				} else {
					cb.successChannel <- val
				}
			case <-ctx.Done():
				js.log.Debug("stopping callback handling loop")
				return
			}
		}
	}()

	return callbackErrChan
}

func (js *nodejsSvc) startRequest(ctx context.Context, funcName string, eventData []byte) error {

	req := pbUserCode.StartRequest{
		EntryPoint: fmt.Sprintf("%s:%s", js.fileName, funcName),
		Event: &pbUserCode.Event{
			Data: eventData,
		},
	}
	if _, err := js.runner.Start(ctx, &req); err != nil {
		// TODO: Handle traceback
		return err
	}

	return nil
}

func (js *nodejsSvc) setupHealthcheck(ctx context.Context) chan (error) {

	runnerHealthChan := make(chan error, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				js.log.Debug("health check loop stopped")
				return
			case <-time.After(10 * time.Second):
				healthReq := pbUserCode.RunnerHealthRequest{}

				resp, err := js.runner.Health(ctx, &healthReq)
				if err != nil { // no network/lost packet.load? for sanity check the state locally via IPC/signals
					err = runnerManager.RunnerHealth(ctx, js.runnerID)
				} else if resp.Error != "" {
					err = fmt.Errorf("grpc: %s", resp.Error)
				}
				if err != nil {
					js.log.Error("runner health failed", zap.Error(err))

					// TODO: ENG-1675 - cleanup runner junk

					runnerHealthChan <- err
					return
				}

			}
		}
	}()
	return runnerHealthChan
}

// initialCall handles initial call from autokitteh, it does the message loop with Python.
// We split it from Call since Call is also used to execute activities.
func (js *nodejsSvc) initialCall(ctx context.Context, funcName string, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(args) > 0 {
		return sdktypes.InvalidValue, errors.New("initial call can't have positional args")
	}

	defer func() {
		js.cleanup(ctx)
		js.log.Info("Python subprocess cleanup after initial call is done")
	}()

	js.log.Info("initial call", zap.Any("func", funcName))
	event, err := js.kwToEvent(kwargs)
	if err != nil {
		js.log.Error("can't convert event", zap.Error(err))
		return sdktypes.InvalidValue, fmt.Errorf("can't convert: %w", err)
	}

	keys := slices.Collect(maps.Keys(event))
	js.log.Info("event", zap.Any("keys", keys))

	eventData, err := json.Marshal(event)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("marshal event: %w", err)
	}

	cancellableCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	callbackErrChan := js.setupCallbacksListeningLoop(cancellableCtx)

	if err := js.startRequest(ctx, funcName, eventData); err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("start request: %w", err)
	}

	runnerHealthChan := js.setupHealthcheck(cancellableCtx)

	// Wait for client Done message
	var done *pbUserCode.DoneRequest
	for {
		select {
		case healthErr := <-runnerHealthChan:
			if healthErr != nil {
				return sdktypes.InvalidValue, sdkerrors.NewRetryableErrorf("runner health: %w", healthErr)
			}
		case callbackErr := <-callbackErrChan:
			return sdktypes.InvalidValue, callbackErr
		case v := <-js.channels.done:
			js.log.Info("done signal", zap.String("error", v.Error))
			pCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			js.drainPrints(pCtx)

			done = v
			if done.Error != "" {
				js.log.Info("done error", zap.String("error", done.Error))
				perr := sdktypes.NewProgramError(
					sdktypes.NewStringValue(done.Error),
					js.tracebackToLocation(done.Traceback),
					map[string]string{"raw": done.Error},
				)
				return sdktypes.InvalidValue, perr.ToError()
			}

			if done.Result == nil {
				js.log.Error("done: nil result")
				return sdktypes.InvalidValue, errors.New("done result is nil")
			}

			done.Result.Custom.ExecutorId = js.xid.String()
			return sdktypes.ValueFromProto(done.Result)
		case <-ctx.Done():
			return sdktypes.InvalidValue, fmt.Errorf("context expired - %w", ctx.Err())
		}
	}
}

// drainPrints drains the print channel at the end of a run.
func (js *nodejsSvc) drainPrints(ctx context.Context) {
	// flush the rest of the prints and logs.
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-js.channels.log:
			js.log.Log(pyLevelToZap(r.level), r.message)
		case r := <-js.channels.print:
			js.cbs.Print(ctx, js.runID, r.message)
			close(r.doneChannel)
		}
	}
}

func (js *nodejsSvc) tracebackToLocation(traceback []*pbUserCode.Frame) []sdktypes.CallFrame {
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
			js.log.Warn("can't translate traceback frame", zap.Error(err))
		}
	}

	return frames
}

// Call handles a function call from autokitteh.
// First used of Call start a workflow, later invocations are activity calls.
func (js *nodejsSvc) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	fn := v.GetFunction()
	if !fn.IsValid() {
		js.log.Error("call - invalid function", zap.Any("function", v))
		return sdktypes.InvalidValue, fmt.Errorf("%#v is not a function", v)
	}

	fnName := fn.Name().String()
	js.log.Info("call", zap.String("func", fnName))

	if js.firstCall {
		js.firstCall = false
		return js.initialCall(ctx, fnName, args, kwargs)
	}

	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			select {
			case r := <-js.channels.log:
				js.log.Log(pyLevelToZap(r.level), r.message)
				close(r.doneChannel)
			case p := <-js.channels.print:
				js.cbs.Print(ctx, js.runID, p.message)
				close(p.doneChannel)
			case <-done:
				return
			}
		}
	}()

	// If we're here, it's an activity call
	req := pbUserCode.ExecuteRequest{
		Data: fn.Data(),
	}
	resp, err := js.runner.Execute(ctx, &req)
	switch {
	case err != nil:
		return sdktypes.InvalidValue, err
	case resp.Error != "":
		js.log.Warn("activity error", zap.String("error", resp.Error))
	}

	resp.Result.Custom.ExecutorId = js.xid.String()
	return sdktypes.ValueFromProto(resp.Result)
}
