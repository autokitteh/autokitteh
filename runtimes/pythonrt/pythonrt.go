package pythonrt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
	pb_value "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/values/v1"
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
	done     chan *pb.DoneRequest
	err      chan string
	request  chan *pb.ActivityRequest
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
	// remote       *workerGRPCHandler

	// runner       Runner
	runner *RunnerClient
	// runnerManager pb.RunnerManagerClient
	runnerID string

	firstCall bool // first call is the trigger, other calls are activities
	accPrints []*logMessage
	accLogs   []*logMessage

	channels comChannels

	syscallFn sdktypes.Value
}

func (py *pySvc) cleanup(ctx context.Context) {
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
		l.Info("remote runner configued")
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
			done:     make(chan *pb.DoneRequest, 1),
			err:      make(chan string, 1),
			request:  make(chan *pb.ActivityRequest, 1),
			print:    make(chan *logMessage, 1),
			log:      make(chan *logMessage, 1),
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
		{Name: "created_at"},
		{Name: "data"},
		{Name: "event_id"},
		{Name: "integration_id"},
	},
}))

func entriesToValues(xid sdktypes.ExecutorID, entries []string) (map[string]sdktypes.Value, error) {
	values := make(map[string]sdktypes.Value)
	for _, name := range entries {
		fn, err := sdktypes.NewFunctionValue(xid, name, nil, nil, pyModuleFunc)
		if err != nil {
			return nil, err
		}
		values[name] = fn
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
		return sdktypes.InvalidValue, fmt.Errorf("`ak` is not a struct")
	}

	syscall, ok := ak.GetStruct().Fields()["syscall"]
	if !ok {
		return sdktypes.InvalidValue, fmt.Errorf("`syscall` not found in `ak`")
	}
	if !syscall.IsFunction() {
		return sdktypes.InvalidValue, fmt.Errorf("`syscall` is not a function")
	}

	return syscall, nil
}

// entryPointFileName strips the handler from the entry point
// "porgram.py:on_event" -> "program.py"
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
func (py *pySvc) Run(
	ctx context.Context,
	runID sdktypes.RunID,
	sessionID sdktypes.SessionID,
	mainPath string,
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	runnerOK := false
	py.ctx = ctx
	py.runID = runID
	py.sessionID = sessionID
	py.xid = sdktypes.NewExecutorID(runID) // Should be first
	py.log = py.log.With(zap.String("run_id", runID.String()), zap.String("session_id", sessionID.String()), zap.String("path", mainPath))

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

	py.syscallFn, err = loadSyscall(values)
	if err != nil {
		return nil, fmt.Errorf("can't load syscall: %w", err)
	}

	runnerID, runner, err := runnerManager.Start(ctx, sessionID, tarData, py.envVars)
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
	req := pb.ExportsRequest{
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
	fn := val.GetFunction()

	reply := func(req *pb.ActivityReplyRequest) {
		if req.Error != "" {
			py.log.Warn("activity reply error", zap.String("error", req.Error))
		}

		req.Data = fn.Data()

		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		reply, err := py.runner.ActivityReply(ctx, req)
		switch {
		case err != nil:
			py.log.Warn("activity reply error", zap.Error(err))
		case reply.Error != "":
			py.log.Warn("activity reply error", zap.String("reply error", reply.Error))
		}
	}

	if !val.IsFunction() {
		reply(&pb.ActivityReplyRequest{Error: fmt.Sprintf("%#v is not a function", val)})
		return
	}

	out, err := py.cbs.Call(py.ctx, py.runID, val, args, kw)
	if err == nil && !out.IsBytes() {
		err = fmt.Errorf("call output not bytes: %#v", out)
	}
	if err != nil {
		reply(&pb.ActivityReplyRequest{Error: err.Error()})
		return
	}

	reply(&pb.ActivityReplyRequest{Result: out.GetBytes().Value()})
}

// initialCall handles initial call from autokitteh, it does the message loop with Python.
// We split it from Call since Call is also used to execute activities.
func (py *pySvc) initialCall(ctx context.Context, funcName string, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	if len(args) > 0 {
		return sdktypes.InvalidValue, fmt.Errorf("initial call can't have positional args")
	}

	defer func() {
		py.cleanup(ctx)
		py.log.Info("Python subprocess cleanup after initial call is done")
	}()

	py.log.Info("initial call", zap.Any("func", funcName))
	event, err := py.kwToEvent(kwargs)
	if err != nil {
		py.log.Error("can't convert event", zap.Error(err))
		return sdktypes.InvalidValue, fmt.Errorf("can't convert: %w", err)
	}

	py.log.Info("event", zap.Any("keys", maps.Keys(event)))

	eventData, err := json.Marshal(event)
	if err != nil {
		return sdktypes.InvalidValue, fmt.Errorf("marshal event: %w", err)
	}

	req := pb.StartRequest{
		EntryPoint: fmt.Sprintf("%s:%s", py.fileName, funcName),
		Event: &pb.Event{
			Data: eventData,
		},
	}
	if _, err := py.runner.Start(ctx, &req); err != nil {
		// TODO: Handle traceback
		return sdktypes.InvalidValue, err
	}

	cancellableCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	runnerHealthChan := make(chan error, 1)
	go func() {
		for {
			select {
			case <-cancellableCtx.Done():
				py.log.Debug("health check loop stopped")
				return
			case <-time.After(10 * time.Second):
				healthReq := pb.HealthRequest{}

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
	}()

	// Wait for client Done message
	var done *pb.DoneRequest
	for {
		py.emitPrintsAndLogs(ctx)

		select {
		case healthErr := <-runnerHealthChan:
			if healthErr != nil {
				return sdktypes.InvalidValue, sdkerrors.NewRetryableError("runner health: %w", healthErr)
			}
		case r := <-py.channels.log:
			py.accLogs = append(py.accLogs, r)
		case r := <-py.channels.print:
			py.accPrints = append(py.accPrints, r)
		case r := <-py.channels.request:
			var (
				fnName = "pyFunc"
				args   []sdktypes.Value
				kw     map[string]sdktypes.Value
			)

			if r.CallInfo != nil {
				fnName = r.CallInfo.Function
				args = kittehs.Transform(r.CallInfo.Args, func(v *pb_value.Value) sdktypes.Value {
					// TODO: What if there's an error
					val, _ := sdktypes.ValueFromProto(v)
					return val
				})
				kw = kittehs.TransformMap(r.CallInfo.Kwargs, func(k string, v *pb_value.Value) (string, sdktypes.Value) {
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
			val, err := py.cbs.Call(ctx, py.runID, py.syscallFn, cb.args, cb.kwargs)
			if err != nil {
				cb.errorChannel <- err
			} else {
				cb.successChannel <- val
			}
		case v := <-py.channels.done:
			pCtx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			py.drainPrints(pCtx)

			done = v
			if done.Error != "" {
				perr := sdktypes.NewProgramError(
					sdktypes.NewStringValue(done.Error),
					py.tracebackToLocation(done.Traceback),
					map[string]string{"raw": done.Error},
				)
				return sdktypes.InvalidValue, perr.ToError()
			}

			return sdktypes.ValueFromProto(done.Result)
		case <-ctx.Done():
			return sdktypes.InvalidValue, fmt.Errorf("context expired - %w", ctx.Err())
		}
	}
}

func (py *pySvc) emitPrintsAndLogs(ctx context.Context) {
	for ; len(py.accPrints) > 0; py.accPrints = py.accPrints[1:] {
		p := py.accPrints[0]
		py.cbs.Print(ctx, py.runID, p.message)

		if p.doneChannel != nil {
			close(p.doneChannel)
		}
	}

	for ; len(py.accLogs) > 0; py.accLogs = py.accLogs[1:] {
		p := py.accLogs[0]
		py.log.Log(pyLevelToZap(p.level), p.message)

		if p.doneChannel != nil {
			close(p.doneChannel)
		}
	}
}

// drainPrints drains the print channel at the end of a run.
func (py *pySvc) drainPrints(ctx context.Context) {
	// emit already accumulated prints and logs.
	py.emitPrintsAndLogs(ctx)

	// flush the rest of the prints and logs.
	for {
		select {
		case <-ctx.Done():
			return
		case r := <-py.channels.log:
			py.log.Log(pyLevelToZap(r.level), r.message)
		case r := <-py.channels.print:
			py.cbs.Print(ctx, py.runID, r.message)
		}
	}
}

func (py *pySvc) tracebackToLocation(traceback []*pb.Frame) []sdktypes.CallFrame {
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
	fn := v.GetFunction()
	if !fn.IsValid() {
		py.log.Error("call - invalid function", zap.Any("function", v))
		return sdktypes.InvalidValue, fmt.Errorf("%#v is not a function", v)
	}

	fnName := fn.Name().String()
	py.log.Info("call", zap.String("func", fnName))

	if py.firstCall {
		py.firstCall = false
		return py.initialCall(ctx, fnName, args, kwargs)
	}

	// If we're here, it's an activity call
	req := pb.ExecuteRequest{
		Data: fn.Data(),
	}
	resp, err := py.runner.Execute(ctx, &req)
	switch {
	case err != nil:
		return sdktypes.InvalidValue, err
	case resp.Error != "":
		return sdktypes.InvalidValue, fmt.Errorf("%s", resp.Error)
	}

	py.emitPrintsAndLogs(ctx)

	return sdktypes.NewBytesValue(resp.Result), nil
}
