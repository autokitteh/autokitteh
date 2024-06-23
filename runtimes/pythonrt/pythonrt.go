package pythonrt

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/xdg"
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

type pySvc struct {
	log       *zap.Logger
	run       *pyRunInfo
	xid       sdktypes.ExecutorID
	cbs       *sdkservices.RunCallbacks
	exports   map[string]sdktypes.Value
	firstCall bool
	comm      *Comm
	stdout    *streamLogger
	stderr    *streamLogger
	syscallFn sdktypes.Value
	pyExe     string
}

var minPyVersion = Version{
	Major: 3,
	Minor: 11,
}

func isGoodVersion(v Version) bool {
	if v.Major < minPyVersion.Major {
		return false
	}

	return v.Minor >= minPyVersion.Minor
}

const exeEnvKey = "AK_WORKER_PYTHON"

func New() (sdkservices.Runtime, error) {
	log, err := logger.New(logger.Configs.Dev) // TODO (ENG-553): From configuration
	if err != nil {
		return nil, err
	}
	log = log.With(zap.String("runtime", "python"))

	svc := pySvc{
		log: log,
	}

	userPython := true
	pyExe := os.Getenv(exeEnvKey)
	if pyExe == "" {
		pyExe, err = findPython()
		if err != nil {
			return nil, err
		}
		userPython = false
	}

	if userPython {
		log.Info("user python", zap.String("python", pyExe))

	}

	const timeout = 3 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	info, err := pyExeInfo(ctx, pyExe)
	if err != nil {
		return nil, err
	}

	log.Info("python info", zap.String("exe", info.Exe), zap.Any("version", info.Version))
	if !isGoodVersion(info.Version) {
		const format = "python >= %d.%d required, found %q"
		return nil, fmt.Errorf(format, minPyVersion.Major, minPyVersion.Minor, info.VersionString)
	}

	// If user supplies which Python to use, we use it "as-is" without creating venv
	if !userPython {
		if err := ensureVEnv(log, pyExe); err != nil {
			return nil, fmt.Errorf("create venv: %w", err)
		}
		svc.pyExe = venvPy
	} else {
		svc.pyExe = pyExe
	}
	log.Info("using python", zap.String("exe", svc.pyExe))

	return &svc, nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func ensureVEnv(log *zap.Logger, pyExe string) error {
	if dirExists(venvPath) {
		return nil
	}

	log.Info("creating venv", zap.String("path", venvPath))
	return createVEnv(pyExe, venvPath)
}

func (*pySvc) Get() sdktypes.Runtime { return Runtime.Desc }

const archiveKey = "archive"

func asBuildExport(e Export) sdktypes.BuildExport {
	pb := sdktypes.BuildExportPB{
		Symbol: e.Name,
		Location: &sdktypes.CodeLocationPB{
			Path: e.File,
			Row:  uint32(e.Line),
			Col:  1,
		},
	}

	b, _ := sdktypes.BuildExportFromProto(&pb)
	return b
}

func (py *pySvc) Build(ctx context.Context, fs fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	py.log.Info("build")

	data, err := createTar(fs)
	if err != nil {
		py.log.Error("create tar", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	exports, err := pyExports(ctx, py.pyExe, fs)
	if err != nil {
		py.log.Error("get exports", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	buildExports := kittehs.Transform(exports, asBuildExport)
	var art sdktypes.BuildArtifact
	art = art.WithCompiledData(
		map[string][]byte{
			archiveKey: data,
		},
	).WithExports(buildExports)

	return art, nil
}

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

func (py *pySvc) handleLog(msg Message) error {
	log, err := extractMessage[LogMessage](msg)
	if err != nil {
		return err
	}

	level := pyLevelToZap(log.Level)
	py.log.Log(level, log.Message, zap.String("source", "python"))
	return nil
}

func (py *pySvc) loadSyscall(values map[string]sdktypes.Value) error {
	ak, ok := values["ak"]
	if !ok {
		return fmt.Errorf("`ak` not found")
	}

	if !ak.IsStruct() {
		return fmt.Errorf("`ak` is not a struct")
	}

	syscall, ok := ak.GetStruct().Fields()["syscall"]
	if !ok {
		return fmt.Errorf("`syscall` not found in `ak`")
	}
	if !syscall.IsFunction() {
		return fmt.Errorf("`syscall` is not a function")
	}

	py.syscallFn = syscall
	return nil
}

func (py *pySvc) handleSleep(ctx context.Context, msg Message) error {
	sleep, err := extractMessage[SleepMessage](msg)
	if err != nil {
		py.log.Error("sleep message", zap.Error(err))
		return err
	}

	py.log.Info("sleep", zap.Float64("seconds", sleep.Seconds))

	// Milliseconds sleep granularity should be good enough.
	d := time.Duration(sleep.Seconds*1000) * time.Millisecond
	args := []sdktypes.Value{
		sdktypes.NewStringValue("sleep"),
		sdktypes.NewDurationValue(d),
	}

	_, err = py.cbs.Call(ctx, py.xid.ToRunID(), py.syscallFn, args, nil)
	if err != nil {
		py.log.Error("call sleep", zap.Error(err))
		return err
	}

	return py.comm.Send(sleep)
}

/*
Run starts a Python workflow.

It'll load the Python module and send back the list of exported names.
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
	py.xid = sdktypes.NewExecutorID(runID) // Should be first
	py.log = py.log.With(zap.String("run_id", runID.String()))
	py.log.Info("run", zap.String("path", mainPath))

	if err := py.loadSyscall(values); err != nil {
		return nil, err
	}

	py.xid = sdktypes.NewExecutorID(runID)
	py.log.Info("executor", zap.String("id", py.xid.String()))
	py.cbs = cbs
	py.firstCall = true // State for Call.

	// Load environment defined by user in the `vars` section of the manifest,
	// these are injected to the Python subprocess environment.
	env, err := cbs.Load(ctx, runID, "env")
	if err != nil {
		return nil, fmt.Errorf("can't load env : %w", err)
	}
	envMap := kittehs.TransformMap(env, func(key string, value sdktypes.Value) (string, string) {
		return key, value.GetString().Value()
	})
	py.log.Info("env", zap.Any("env", envMap))

	tarData := compiled[archiveKey]
	if tarData == nil {
		return nil, fmt.Errorf("%q note found in compiled data", archiveKey)
	}

	py.stdout = newStreamLogger("[stdout] ", cbs.Print, runID)
	py.stderr = newStreamLogger("[stderr] ", cbs.Print, runID)
	opts := runOptions{
		log:      py.log,
		pyExe:    py.pyExe,
		tarData:  tarData,
		rootPath: mainPath,
		env:      envMap,
		stdout:   py.stdout,
		stderr:   py.stderr,
	}
	ri, err := runPython(opts)
	if err != nil {
		return nil, err
	}
	py.run = ri

	// Kill Python process in case we had errors.
	killPy := true
	defer func() {
		if !killPy {
			return
		}

		py.log.Error("killing Python", zap.Int("pid", py.run.proc.Pid))
		if err := py.run.proc.Kill(); err != nil {
			py.log.Warn("kill", zap.Int("pid", py.run.proc.Pid), zap.Error(err))
		}
		py.run.Cleanup()
	}()

	conn, err := py.run.lis.Accept()
	if err != nil {
		py.log.Error("connect to socket", zap.Error(err))
		return nil, err
	}
	py.log.Info("python connected", zap.String("peer", conn.RemoteAddr().String()))
	py.comm = NewComm(conn)

	// FIXME (ENG-577) We might get activity calls before module is loaded if there are module level function calls.
	var mod ModuleMessage
	for {
		msg, err := py.comm.Recv()
		if err != nil {
			py.log.Error("initial message from python", zap.Error(err))
			return nil, err
		}

		if msg.Type == "" {
			py.log.Error("python can't load module", zap.String("module", mainPath))
			return nil, fmt.Errorf("python can't load %q", mainPath)
		}

		if msg.Type == messageType[LogMessage]() {
			if err := py.handleLog(msg); err != nil {
				return nil, err
			}
			continue
		}

		mod, err = extractMessage[ModuleMessage](msg)
		if err != nil {
			py.log.Error("initial message from python", zap.Error(err))
			return nil, err
		}

		py.log.Info("module loaded")
		break
	}

	py.log.Info("module entries", zap.Any("entries", mod.Entries))
	exports, err := entriesToValues(py.xid, mod.Entries)
	if err != nil {
		py.log.Error("can't create module entries", zap.Error(err))
		return nil, fmt.Errorf("can't create module entries: %w", err)
	}
	py.exports = exports

	killPy = false // All is good, don't kill Python subprocess.

	py.log.Info("run done")
	return py, nil
}

func (py *pySvc) ID() sdktypes.RunID              { return py.xid.ToRunID() }
func (py *pySvc) ExecutorID() sdktypes.ExecutorID { return py.xid }

func (py *pySvc) Values() map[string]sdktypes.Value {
	return py.exports
}

func (py *pySvc) Close() {
	py.log.Info("closing (not really)")
	// AK calls Close after `Run`, but we need the Python process running for `Call` as well.
	// We kill the Python process once the initial `Call` is completed.
}

// initialCall handles initial call from autokitteh.
// We split it from Call since Call is also used to execute activities.
func (py *pySvc) initialCall(ctx context.Context, funcName string, event map[string]any) (sdktypes.Value, error) {
	defer func() {
		py.log.Info("python done, killing")

		py.stderr.Close()
		py.stdout.Close()
		py.comm.Close()

		if py.run.proc != nil {
			if err := py.run.proc.Kill(); err != nil {
				py.log.Warn("kill", zap.Int("pid", py.run.proc.Pid), zap.Error(err))
			}
		}
		py.run.Cleanup()
		py.run.proc = nil
	}()

	py.log.Info("initial call", zap.Any("func", funcName))

	// Initial run call.
	msg := RunMessage{
		FuncName: funcName,
		Event:    event,
	}
	if err := py.comm.Send(msg); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Activity callback loop.
	for {
		py.log.Info("waiting for Python call")
		msg, err := py.comm.Recv()
		if err != nil {
			py.log.Error("communication error", zap.Error(err))
			return sdktypes.InvalidValue, err
		}
		py.log.Debug("from python", zap.Any("message", msg))

		if msg.Type == "" {
			py.log.Error("empty message from python, probably error", zap.Any("message", msg))
			return sdktypes.InvalidValue, fmt.Errorf("empty message from Python")
		}

		if msg.Type == messageType[DoneMessage]() {
			break
		}

		if msg.Type == messageType[LogMessage]() {
			if err := py.handleLog(msg); err != nil {
				return sdktypes.InvalidValue, err
			}
			continue
		}

		if msg.Type == messageType[SleepMessage]() {
			if err := py.handleSleep(ctx, msg); err != nil {
				return sdktypes.InvalidValue, err
			}
			continue
		}

		cbm, err := extractMessage[CallbackMessage](msg)
		if err != nil {
			py.log.Error("callback", zap.Error(err))
			return sdktypes.InvalidValue, err
		}

		// Generate activity, it'll call Python with the result
		// The function name is irrelevant, all the information Python needs is in the Payload
		fn, err := sdktypes.NewFunctionValue(py.xid, cbm.Name, cbm.Data, nil, pyModuleFunc)
		if err != nil {
			py.log.Error("create function", zap.Error(err))
			return sdktypes.InvalidValue, err
		}

		py.log.Info("callback", zap.String("func", cbm.Name))
		val, err := py.cbs.Call(
			ctx,
			py.xid.ToRunID(),
			// The Python function to call is encoded in the payload
			fn,
			kittehs.Transform(cbm.Args, sdktypes.NewStringValue),
			kittehs.TransformMap(cbm.Kw, func(key, val string) (string, sdktypes.Value) {
				return key, sdktypes.NewStringValue(val)
			}),
		)
		if err != nil {
			py.log.Error("callback", zap.Error(err))
			return sdktypes.InvalidValue, err
		}

		if !val.IsBytes() {
			py.log.Error("activity result should be bytes", zap.Any("value", val))
			return sdktypes.InvalidValue, err
		}

		reply := ResponseMessage{
			Value: val.GetBytes().Value(),
		}
		if err := py.comm.Send(reply); err != nil {
			py.log.Error("send value to Python", zap.Error(err))
			return sdktypes.InvalidValue, err
		}
	}

	return sdktypes.Nothing, nil
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

	out, err := cbs.Call(ctx, py.xid.ToRunID(), fn, nil, nil)
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

// Call handles a function call from autokitteh.
// First used of Call start a workflow, later invocations are activity calls.
func (py *pySvc) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	py.log.Info("call", zap.String("func", v.String()))
	if py.run.proc == nil {
		py.log.Error("call - python not running")
		return sdktypes.InvalidValue, fmt.Errorf("python not running")
	}

	fn := v.GetFunction()
	if !fn.IsValid() {
		py.log.Error("call - invalid function", zap.Any("function", v))
		return sdktypes.InvalidValue, fmt.Errorf("%#v is not a function", v)
	}

	// Convert event to JSON
	event := make(map[string]any, len(kwargs))
	for key, val := range kwargs {
		goVal, err := unwrap(val)
		if err != nil {
			return sdktypes.InvalidValue, err
		}
		event[key] = goVal
	}
	py.log.Info("event", zap.Any("event", event))

	if err := py.injectHTTPBody(ctx, kwargs["data"], event, py.cbs); err != nil {
		return sdktypes.InvalidValue, err
	}

	fnName := fn.Name().String()
	py.log.Info("call", zap.String("function", fnName))
	if py.firstCall { // TODO: mutex. Ask Itay
		py.firstCall = false

		return py.initialCall(ctx, fnName, event)
	}

	// Activity call
	cbm := CallbackMessage{
		Name: fnName,
		Data: fn.Data(),
	}
	py.log.Info("callback to Python", zap.Any("message", cbm))

	if err := py.comm.Send(cbm); err != nil {
		py.log.Error("send to python", zap.Error(err))
		return sdktypes.InvalidValue, err
	}

	var msg Message
	for { // Consume logs.
		var err error
		msg, err = py.comm.Recv()
		if err != nil {
			py.log.Error("from python", zap.Error(err))
			return sdktypes.InvalidValue, err
		}

		if messageType[LogMessage]() != msg.Type {
			break
		}

		if err := py.handleLog(msg); err != nil {
			py.log.Error("handle log", zap.Error(err))
			return sdktypes.InvalidValue, err
		}
	}

	rm, err := extractMessage[ResponseMessage](msg)
	if err != nil {
		py.log.Error("from python", zap.Error(err))
		return sdktypes.InvalidValue, err
	}
	py.log.Info("python return", zap.Any("message", rm))

	return sdktypes.NewBytesValue(rm.Value), nil
}
