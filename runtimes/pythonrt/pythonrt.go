package pythonrt

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
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

func New() (sdkservices.Runtime, error) {
	// Use sdklogger
	log, err := logger.New(logger.Configs.Dev) // TODO (ENG-553): From configuration
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	info, err := pyExeInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("python info: %w", err)
	}

	log.Info("system python info", zap.String("exe", info.Exe), zap.Any("version", info.Version))
	if !isGoodVersion(info.Version) {
		const format = "python >= %d.%d required, found %q"
		return nil, fmt.Errorf(format, minPyVersion.Major, minPyVersion.Minor, info.VersionString)
	}

	if err := ensureVEnv(log, info.Exe); err != nil {
		return nil, fmt.Errorf("create venv: %w", err)
	}

	log.Info("venv python", zap.String("exe", venvPy))

	svc := pySvc{
		log: log,
	}

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

func (py *pySvc) Build(ctx context.Context, fs fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	py.log.Info("build")

	data, err := createTar(fs)
	if err != nil {
		py.log.Error("create tar", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	var art sdktypes.BuildArtifact
	art = art.WithCompiledData(
		map[string][]byte{
			archiveKey: data,
		},
	)

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
	py.log.Info("run", zap.String("id", runID.String()), zap.String("path", mainPath))

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

	ri, err := runPython(py.log, venvPy, tarData, mainPath, envMap)
	if err != nil {
		return nil, err
	}

	// Kill Python process in case we had errors.
	killPy := true
	defer func() {
		if !killPy {
			return
		}

		py.log.Error("killing Python", zap.Int("pid", ri.proc.Pid))
		if err := ri.proc.Kill(); err != nil {
			py.log.Warn("kill", zap.Int("pid", ri.proc.Pid), zap.Error(err))
		}
	}()

	conn, err := ri.lis.Accept()
	if err != nil {
		py.log.Error("connect to socket", zap.Error(err))
		return nil, err
	}
	py.log.Info("python connected", zap.String("peer", conn.RemoteAddr().String()))
	comm := NewComm(conn)

	// FIXME (ENG-577) We might get activity calls before module is loaded if there are module level function calls.

	msg, err := comm.Recv()
	if err != nil {
		py.log.Error("initial message from python", zap.Error(err))
		return nil, err
	}

	if msg.Type == "" {
		py.log.Error("python can't load module", zap.String("module", mainPath))
		return nil, fmt.Errorf("python can't load %q", mainPath)
	}

	mod, err := extractMessage[ModuleMessage](msg)
	if err != nil {
		py.log.Error("initial message from python", zap.Error(err))
		return nil, err
	}

	py.log.Info("module loaded")

	xid := sdktypes.NewExecutorID(runID)
	py.log.Info("executor", zap.String("id", xid.String()))

	py.log.Info("module entries", zap.Any("entries", mod.Entries))
	exports, err := entriesToValues(xid, mod.Entries)
	if err != nil {
		py.log.Error("can't create module entries", zap.Error(err))
		return nil, fmt.Errorf("can't create module entries: %w", err)
	}

	killPy = false // All is good, don't kill Python subprocess.

	py.cbs = cbs
	py.comm = comm
	py.exports = exports
	py.firstCall = true
	py.run = ri
	py.xid = xid

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

func (py *pySvc) handleLog(msg Message) error {
	log, err := extractMessage[LogMessage](msg)
	if err != nil {
		return err
	}

	var fn func(string, ...zap.Field)
	switch log.Level {
	case "DEBUG":
		fn = py.log.Debug
	case "INFO":
		fn = py.log.Info
	case "WARN":
		fn = py.log.Warn
	case "ERROR":
		fn = py.log.Error
	default:
		return fmt.Errorf("unknown log level in %#v", log)
	}

	fn(log.Message, zap.String("runtime", "python"), zap.String("type", "log"))
	return nil
}

// initialCall handles initial call from autokitteh.
// We split it from Call since Call is also used to execute activities.
func (py *pySvc) initialCall(ctx context.Context, funcName string, event map[string]any) (sdktypes.Value, error) {
	defer func() {
		py.log.Info("python done, killing")
		py.comm.Close()
		if err := py.run.proc.Kill(); err != nil {
			py.log.Warn("kill", zap.Int("pid", py.run.proc.Pid), zap.Error(err))
		}
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

		if msg.Type == MessageType[DoneMessage]() {
			break
		}

		if msg.Type == MessageType[LogMessage]() {
			if err := py.handleLog(msg); err != nil {
				py.log.Error("handle log", zap.Error(err))
				return sdktypes.InvalidValue, fmt.Errorf("bad log message: %w", err)
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
	py.log.Info("call", zap.String("func", v.String()), zap.Any("args", args), zap.Any("kwargs", kwargs))
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

	for {
		msg, err := py.comm.Recv()
		if err != nil {
			py.log.Error("from python", zap.Error(err))
			return sdktypes.InvalidValue, err
		}

		if msg.Type == MessageType[LogMessage]() {
			if err := py.handleLog(msg); err != nil {
				py.log.Error("handle log", zap.Error(err))
				return sdktypes.InvalidValue, err
			}
			continue
		}

		rm, err := extractMessage[ResponseMessage](msg)
		if err != nil {
			py.log.Error("from python", zap.Error(err))
			return sdktypes.InvalidValue, err
		}
		py.log.Info("python return", zap.Any("message", rm))

		return sdktypes.NewBytesValue(rm.Value), nil
	}
}
