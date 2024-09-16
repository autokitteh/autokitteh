package pythonrt

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"golang.org/x/exp/maps"

	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	Runtime = &sdkruntimes.Runtime{
		Desc: kittehs.Must1(sdktypes.StrictRuntimeFromProto(&sdktypes.RuntimePB{
			Name:           "python",
			FileExtensions: []string{"py"},
		})),
		New: New,
	}
	venvPath = path.Join(xdg.DataHomeDir(), "venv")
	venvPy   = path.Join(venvPath, "bin", "python")
)

type comChannels struct {
	done    chan *pb.DoneRequest
	err     chan string
	request chan *pb.ActivityRequest
	print   chan *pb.PrintRequest
	log     chan *pb.LogRequest
}

type pySvc struct {
	ctx      context.Context
	log      *zap.Logger
	xid      sdktypes.ExecutorID
	runID    sdktypes.RunID
	cbs      *sdkservices.RunCallbacks
	exports  map[string]sdktypes.Value
	fileName string // main user code file name (entry point)
	// remote       *workerGRPCHandler

	// runner       Runner
	runner pb.RunnerClient
	// runnerManager pb.RunnerManagerClient
	runnerID string

	firstCall bool // first call is the trigger, other calls are activities

	channels comChannels

	syscallFn sdktypes.Value
}

func (py *pySvc) cleanup(ctx context.Context) {
	if err := runnerManager.Stop(ctx, py.runnerID); err != nil {
		py.log.Warn("close runner", zap.Error(err))
	}

	if err := removeRunnerFromServer(py.runnerID); err != nil {
		py.log.Warn("remove runner from grpc", zap.Error(err))
	}
}

func New() (sdkservices.Runtime, error) {
	log, err := logger.New(logger.Configs.Dev) // TODO (ENG-553): From configuration
	if err != nil {
		return nil, err
	}
	log = log.With(zap.String("runtime", "python"))

	svc := pySvc{
		log:       log,
		firstCall: true,
		channels: comChannels{
			done:    make(chan *pb.DoneRequest, 1),
			err:     make(chan string, 1),
			request: make(chan *pb.ActivityRequest, 1),
			print:   make(chan *pb.PrintRequest, 1),
			log:     make(chan *pb.LogRequest, 1),
		},
	}

	return &svc, nil
}

func (*pySvc) Get() sdktypes.Runtime { return Runtime.Desc }

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
	mainPath string,
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	runnerOK := false
	py.ctx = ctx
	py.runID = runID
	py.xid = sdktypes.NewExecutorID(runID) // Should be first
	py.log = py.log.With(zap.String("run_id", runID.String()))
	py.log.Info("run", zap.String("path", mainPath))

	py.log.Info("executor", zap.String("id", py.xid.String()))
	py.cbs = cbs

	// Load environment defined by user in the `vars` section of the manifest,
	// these are injected to the Python subprocess environment.
	env, err := cbs.Load(ctx, runID, "env")
	if err != nil {
		return nil, fmt.Errorf("can't load env : %w", err)
	}
	envMap := kittehs.TransformMap(env, func(key string, value sdktypes.Value) (string, string) {
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

	runnerID, runner, err := runnerManager.Start(ctx, tarData, envMap)
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

	// runnerAddr := fmt.Sprintf("localhost:%d", runner.port)
	// py.runnerClient, err = dialRunner(runnerAddr)
	// if err != nil {
	// 	return nil, err
	// }

	py.fileName = entryPointFileName(mainPath)
	req := pb.ExportsRequest{
		FileName: py.fileName,
	}

	resp, err := py.runner.Exports(ctx, &req)
	if err != nil {
		return nil, err
	}

	py.log.Info("module entries", zap.Any("entries", resp.Exports))
	exports, err := entriesToValues(py.xid, resp.Exports)
	if err != nil {
		py.log.Error("can't create module entries", zap.Error(err))
		return nil, fmt.Errorf("can't create module entries: %w", err)
	}
	py.exports = exports

	runnerOK = true // All is good, don't kill Python subprocess.

	py.log.Info("run done")
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

// TODO (ENG-624) Remove this once we support callbacks to autokitteh
func (py *pySvc) injectHTTPBody(ctx context.Context, data sdktypes.Value, event map[string]any, cbs *sdkservices.RunCallbacks) error {
	if !data.IsStruct() {
		return nil
	}

	evtData, ok := event["data"]
	if !ok {
		return nil
	}

	fields := data.GetStruct().Fields()
	body, ok := fields["body"]
	if !ok {
		return nil
	}

	if !body.IsStruct() {
		return nil
	}

	fields = body.GetStruct().Fields()
	fn, ok := fields["bytes"]
	if !ok || !fn.IsFunction() {
		return nil
	}

	out, err := cbs.Call(ctx, py.runID, fn, nil, nil)
	if err != nil {
		return err
	}

	if !out.IsBytes() {
		return nil
	}

	bodyData := out.GetBytes().Value()

	m, ok := evtData.(map[string]any)
	if !ok {
		return nil
	}

	m["body"] = bodyData
	return nil
}

func (py *pySvc) kwToEvent(ctx context.Context, kwargs map[string]sdktypes.Value) (map[string]any, error) {
	// Convert event to JSON
	event := make(map[string]any, len(kwargs))
	for key, val := range kwargs {
		goVal, err := unwrap(val)
		if err != nil {
			return nil, err
		}
		event[key] = goVal
	}
	py.log.Info("event", zap.Any("keys", maps.Keys(event)))

	if err := py.injectHTTPBody(ctx, kwargs["data"], event, py.cbs); err != nil {
		return nil, err
	}

	return event, nil
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

func (py *pySvc) call(val sdktypes.Value) {
	req := pb.ActivityReplyRequest{}

	// We want to send reply in any case
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		reply, err := py.runner.ActivityReply(ctx, &req)
		switch {
		case err != nil:
			py.log.Error("activity reply error", zap.Error(err))
		case reply.Error != "":
			py.log.Error("activity reply error", zap.String("error", reply.Error))
		}
	}()

	if !val.IsFunction() {
		py.log.Error("bad function", zap.Any("val", val))
		req.Error = fmt.Sprintf("%#v is not a function", val)
		return
	}

	fn := val.GetFunction()
	req.Data = fn.Data()
	out, err := py.cbs.Call(py.ctx, py.runID, val, nil, nil)

	switch {
	case err != nil:
		req.Error = fmt.Sprintf("%s - %s", fn.Name().String(), err)
		py.log.Error("activity reply error", zap.Error(err))
	case !out.IsBytes():
		req.Error = fmt.Sprintf("call output not bytes: %#v", out)
		py.log.Error("activity reply error", zap.String("error", req.Error))
	default:
		data := out.GetBytes().Value()
		req.Result = data
	}
}

// initialCall handles initial call from autokitteh, it does the message loop with Python.
// We split it from Call since Call is also used to execute activities.
func (py *pySvc) initialCall(ctx context.Context, funcName string, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	defer func() {
		py.cleanup(ctx)
		py.log.Info("Python subprocess cleanup after initial call is done")
	}()

	py.log.Info("initial call", zap.Any("func", funcName))
	event, err := py.kwToEvent(ctx, kwargs)
	if err != nil {
		py.log.Error("can't convert event", zap.Error(err))
		return sdktypes.InvalidValue, fmt.Errorf("can't convert: %w", err)
	}

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

	// Wait for client Done message
	var done *pb.DoneRequest
	for {
		select {
		case r := <-py.channels.log:
			level := pyLevelToZap(r.Level)
			py.log.Log(level, r.Message, zap.String("source", "python"))
		case r := <-py.channels.print:
			py.cbs.Print(ctx, py.runID, r.Message)
		case r := <-py.channels.request:
			fnName := r.CallInfo.Function
			py.log.Info("activity", zap.String("function", fnName))
			// it was already checked before we got here
			fn, _ := sdktypes.NewFunctionValue(py.xid, fnName, r.Data, nil, pyModuleFunc)
			py.call(fn)
		case v := <-py.channels.done:
			done = v
			if done.Error != "" {
				perr := sdktypes.NewProgramError(
					sdktypes.NewStringValue(done.Error),
					py.tracebackToLocation(done.Traceback),
					map[string]string{"raw": done.Error},
				)
				return sdktypes.InvalidValue, perr.ToError()
			}

			return sdktypes.NewBytesValue(done.Result), nil

		case <-ctx.Done():
			return sdktypes.InvalidValue, fmt.Errorf("context expired - %w", ctx.Err())
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

	return sdktypes.NewBytesValue(resp.Result), nil
}
