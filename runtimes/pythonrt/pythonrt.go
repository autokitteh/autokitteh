package pythonrt

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"go.uber.org/zap"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/xdg"
	"go.autokitteh.dev/autokitteh/runtimes/pythonrt/pb"
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
	runID     sdktypes.RunID
	cbs       *sdkservices.RunCallbacks
	exports   map[string]sdktypes.Value
	stdout    *streamLogger
	stderr    *streamLogger
	pyExe     string
	fileName  string // main user code file name (entry point)
	remote    *remoteSvc
	runner    pb.RunnerClient
	firstCall bool // first call is the trigger, other calls are activities
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
		log:       log,
		firstCall: true,
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

func (py *pySvc) Build(ctx context.Context, fsys fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
	py.log.Info("build Python module", zap.String("path", path))

	ffs, err := kittehs.NewFilterFS(fsys, func(entry fs.DirEntry) bool {
		return !strings.Contains(entry.Name(), "__pycache__")
	})
	if err != nil {
		return sdktypes.InvalidBuildArtifact, err
	}

	data, err := createTar(ffs)
	if err != nil {
		py.log.Error("create tar", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}

	compiledData := map[string][]byte{
		archiveKey: data,
	}

	// UI requires file names in the compiled data.
	tf := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tf.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			py.log.Error("next tar", zap.Error(err))
			return sdktypes.InvalidBuildArtifact, err
		}

		if !strings.HasSuffix(hdr.Name, ".py") {
			continue
		}

		compiledData[hdr.Name] = nil
	}

	/* TODO: remove?
	exports, err := pyExports(ctx, py.pyExe, fsys)
	if err != nil {
		py.log.Error("get exports", zap.Error(err))
		return sdktypes.InvalidBuildArtifact, err
	}


	buildExports := kittehs.Transform(exports, asBuildExport)
	*/
	var art sdktypes.BuildArtifact
	art = art.WithCompiledData(compiledData) //.WithExports(buildExports)

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

func (py *pySvc) loadSyscall(values map[string]sdktypes.Value) (sdktypes.Value, error) {
	ak, ok := values["ak"]
	if !ok {
		py.log.Warn("can't find `ak` in values")
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

func dialRunner(addr string) (pb.RunnerClient, error) {
	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	c := pb.NewRunnerClient(conn)
	return c, nil
}

func entryPointFileName(entryPoint string) string {
	i := strings.Index(entryPoint, ":")
	if i > 0 {
		return entryPoint[:i]
	}

	return entryPoint
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
	runnerOK := false
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

	py.stdout = newStreamLogger("[stdout] ", cbs.Print, runID)
	py.stderr = newStreamLogger("[stderr] ", cbs.Print, runID)

	syscallFn, err := py.loadSyscall(values)
	if err != nil {
		return nil, fmt.Errorf("can't load syscall: %w", err)
	}

	rsvc := newRemoteSvc(py.log, cbs, runID, syscallFn)

	if err := rsvc.Start(); err != nil {
		return nil, fmt.Errorf("can't start remote service: %w", err)
	}
	py.remote = rsvc
	defer func() {
		if !runnerOK {
			rsvc.Stop()
		}
	}()

	opts := runOptions{
		log:        py.log,
		pyExe:      py.pyExe,
		entryPoint: mainPath,
		env:        envMap,
		stdout:     py.stdout,
		stderr:     py.stderr,
		tarData:    tarData,
		workerAddr: fmt.Sprintf("localhost:%d", rsvc.port),
	}
	ri, err := runPython(opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := ri.proc.Kill(); err != nil {
			py.log.Warn("kill runner", zap.Int("pid", ri.proc.Pid), zap.Error(err))
		}
		if !runnerOK {
			ri.Cleanup()
		}
	}()

	runnerAddr := fmt.Sprintf("localhost:%d", ri.port)
	py.runner, err = dialRunner(runnerAddr)
	if err != nil {
		return nil, err
	}

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

// initialCall handles initial call from autokitteh, it does the message loop with Python.
// We split it from Call since Call is also used to execute activities.
func (py *pySvc) initialCall(ctx context.Context, funcName string, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	defer func() {
		py.log.Info("Python subprocess cleanup after initial call is done")

		py.stderr.Close()
		py.stdout.Close()

		if py.run.proc != nil {
			if err := py.run.proc.Kill(); err != nil {
				py.log.Warn("kill", zap.Int("pid", py.run.proc.Pid), zap.Error(err))
			}
		}
		py.run.Cleanup()
		py.run.proc = nil
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
		RunId:      py.runID.String(),
		EntryPoint: fmt.Sprintf("%s:%s", py.fileName, funcName),
		Event: &pb.Event{
			Data: eventData,
		},
	}
	if _, err := py.runner.Start(ctx, &req); err != nil {
		// TODO: Handle traceback
		return sdktypes.InvalidValue, err
	}

	select {
	case out := <-py.remote.result:
		return sdktypes.NewBytesValue(out), nil
	case err := <-py.remote.error:
		// TODO: traceback
		return sdktypes.InvalidValue, fmt.Errorf("%s", err)
	case <-ctx.Done():
		return sdktypes.InvalidValue, fmt.Errorf("timeout - %w", ctx.Err())
	}
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

	if py.run.proc == nil {
		py.log.Error("call - python not running")
		return sdktypes.InvalidValue, fmt.Errorf("python not running")
	}

	if py.firstCall {
		py.firstCall = false
		return py.initialCall(ctx, fnName, args, kwargs)
	}

	// If we're here, it's an activity call

}

/*
func packSyscallArgs(msg CallMessage) ([]sdktypes.Value, map[string]sdktypes.Value, error) {
	args := make([]sdktypes.Value, len(msg.Args)+1)
	args[0] = sdktypes.NewStringValue(msg.FuncName)
	// Can't use kittehs.Transform since we want to handle errors.
	for i, arg := range msg.Args {
		var err error
		args[i+1], err = sdktypes.WrapValue(arg)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: unsupported argument: %#v - %w", msg.FuncName, arg, err)
		}
	}

	kw := make(map[string]sdktypes.Value, len(msg.Kw))
	for key, val := range msg.Kw {
		v, err := sdktypes.WrapValue(val)
		if err != nil {
			return nil, nil, fmt.Errorf("%s: unsupported kwarg: %s:%#v - %w", msg.FuncName, key, val, err)
		}
		kw[key] = v
	}

	return args, kw, nil
}
*/
