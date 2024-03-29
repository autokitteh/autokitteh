package pythonrt

import (
	"context"
	"encoding/json"
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

type pySVC struct {
	log       *zap.Logger
	run       *pyRunInfo
	xid       sdktypes.ExecutorID
	cbs       *sdkservices.RunCallbacks
	exports   map[string]sdktypes.Value
	firstCall bool
	dec       *json.Decoder
	enc       *json.Encoder
}

func New() (sdkservices.Runtime, error) {
	log, err := logger.New(logger.Configs.Dev) // TODO: From configuration (ENG-553)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	info, err := pyExecInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("python info - %w", err)
	}

	log.Info("system python info", zap.String("exe", info.Exe), zap.String("version", info.Version))
	if err := ensureVEnv(log, info.Exe); err != nil {
		return nil, fmt.Errorf("create venv - %w", err)
	}

	log.Info("venv python", zap.String("exe", venvPy))

	svc := pySVC{
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

func (*pySVC) Get() sdktypes.Runtime { return Runtime.Desc }

const archiveKey = "archive"

func (py *pySVC) Build(ctx context.Context, fs fs.FS, path string, values []sdktypes.Symbol) (sdktypes.BuildArtifact, error) {
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

type PyMessage struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Payload []byte `json:"payload"`
	Func    struct {
		Name string            `json:"name"`
		Args []string          `json:"args"`
		Kw   map[string]string `json:"kw"`
	} `json:"func"`
}

// All Python handler function get all event information.
var pyModuleFunc = kittehs.Must1(sdktypes.ModuleFunctionFromProto(&sdktypes.ModuleFunctionPB{
	Input: []*sdktypes.ModuleFunctionFieldPB{
		{Name: "event_type"},
		{Name: "event_id"},
		{Name: "original_event_id"},
		{Name: "integration_id"},
		{Name: "data"},
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
Flow (backend/internal/sessions/sessionworkflows/workflow.go)
- Run should return a run
- It should have mainPath in Values as FunctionValue
- AK will call run.Call
-
*/

func (py *pySVC) Run(
	ctx context.Context,
	runID sdktypes.RunID,
	mainPath string,
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {
	py.log.Info("run", zap.String("id", runID.String()), zap.String("path", mainPath))

	env, err := cbs.Load(ctx, runID, "env")
	if err != nil {
		return nil, fmt.Errorf("can't load env - %w", err)
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

	// Cleanup in case we had errors
	killPy := true
	defer func() {
		if killPy {
			py.log.Error("killing Python", zap.Int("pid", ri.proc.Pid))
			if err := ri.proc.Kill(); err != nil {
				py.log.Warn("kill", zap.Int("pid", ri.proc.Pid), zap.Error(err))
			}
		}
	}()

	conn, err := ri.lis.Accept()
	if err != nil {
		py.log.Error("connect to socket", zap.Error(err))
		return nil, err
	}
	py.log.Info("python connected", zap.String("peer", conn.RemoteAddr().String()))

	dec := json.NewDecoder(conn)
	var msg PyMessage
	if err := dec.Decode(&msg); err != nil {
		py.log.Error("initial message from python", zap.Error(err))
		return nil, err
	}

	// FIXME (ENG-577) We might get activity calls before module is loaded if there are module level function calls.
	if msg.Type != "module" {
		py.log.Error("wrong initial message type from python", zap.String("type", msg.Type))
		return nil, fmt.Errorf("wrong initial message: type=%q", msg.Type)
	}
	py.log.Info("module loaded")

	py.xid = sdktypes.NewExecutorID(runID)
	py.log.Info("executor", zap.String("id", py.xid.String()))

	var entries []string
	if err := json.Unmarshal(msg.Payload, &entries); err != nil {
		py.log.Error("can't parse module entries", zap.Error(err))
		return nil, fmt.Errorf("can't parse module entries: %w", err)
	}
	py.log.Info("module entries", zap.Any("entries", entries))

	exports, err := entriesToValues(py.xid, entries)
	if err != nil {
		py.log.Error("can't create module entries", zap.Error(err))
		return nil, fmt.Errorf("can't create module entries: %w", err)
	}

	killPy = false
	py.run = ri
	py.cbs = cbs
	py.exports = exports
	py.firstCall = true
	py.dec = dec
	py.enc = json.NewEncoder(conn)

	return py, nil
}

func (py *pySVC) ID() sdktypes.RunID              { return py.xid.ToRunID() }
func (py *pySVC) ExecutorID() sdktypes.ExecutorID { return py.xid }

func (py *pySVC) Values() map[string]sdktypes.Value {
	return py.exports
}

func (py *pySVC) Close() {
	py.log.Info("closing")
	// AK calls Close after `Run`, but we need the Python process running for `Call` as well.
	// We kill the Python process once the initial `Call` is completed.
}

func (py *pySVC) initialCall(ctx context.Context, funcName string, payload []byte) (sdktypes.Value, error) {
	defer func() {
		py.log.Info("python done, killing")
		if err := py.run.proc.Kill(); err != nil {
			py.log.Warn("kill", zap.Int("pid", py.run.proc.Pid), zap.Error(err))
		}
		py.run.proc = nil
	}()

	// Initial run cal
	msg := PyMessage{
		Type:    "run",
		Name:    funcName,
		Payload: payload,
	}
	py.log.Info("initial call", zap.Any("message", msg))
	if err := py.enc.Encode(msg); err != nil {
		return sdktypes.InvalidValue, err
	}

	// Callbacks
	for {
		var msg PyMessage
		py.log.Info("waiting for Python call")
		if err := py.dec.Decode(&msg); err != nil {
			py.log.Error("communication error", zap.Error(err))
			return sdktypes.InvalidValue, err
		}
		py.log.Info("from python", zap.Any("message", msg))

		if msg.Type == "done" {
			break
		}

		// Generate activity, it'll call Python with the result
		// The function name is irrelevant, all the information Python needs is in the Payload
		fn, err := sdktypes.NewFunctionValue(py.xid, msg.Func.Name, msg.Payload, nil, pyModuleFunc)
		if err != nil {
			py.log.Error("create function", zap.Error(err))
			return sdktypes.InvalidValue, err
		}

		py.log.Info("callback")
		_, err = py.cbs.Call(
			ctx,
			py.xid.ToRunID(),
			// The Python function to call is encoded in the payload
			fn,
			kittehs.Transform(msg.Func.Args, sdktypes.NewStringValue),
			kittehs.TransformMap(msg.Func.Kw, func(key, val string) (string, sdktypes.Value) {
				return key, sdktypes.NewStringValue(val)
			}),
		)
		if err != nil {
			py.log.Error("callback", zap.Error(err))
			return sdktypes.InvalidValue, err
		}
	}

	return sdktypes.Nothing, nil
}

// TODO: We just want the JSON of kwargs, this seems excessive. Ask Itay
func valueToGo(v sdktypes.Value) (any, error) {
	switch {
	case v.IsBoolean():
		return v.GetBoolean().Value(), nil
	case v.IsBytes():
		return v.GetBytes().Value(), nil
	case v.IsDict():
		d := v.GetDict()
		items := d.Items()
		m := make(map[string]any, len(items))
		for _, i := range items {
			if !i.K.IsString() {
				return nil, fmt.Errorf("non string key: %#v", i.K)
			}
			gv, err := valueToGo(i.V)
			if err != nil {
				return nil, err
			}
			m[i.K.GetString().Value()] = gv
		}
		return m, nil
	case v.IsDuration():
		return v.GetDuration().Value(), nil
	case v.IsFloat():
		return v.GetFloat().Value(), nil
	case v.IsInteger():
		return v.GetInteger().Value(), nil
	case v.IsList():
		values := v.GetList().Values()
		lst := make([]any, len(values))
		for i, v := range values {
			gv, err := valueToGo(v)
			if err != nil {
				return nil, err
			}
			lst[i] = gv
		}
	case v.IsNothing():
		return nil, nil
	case v.IsString():
		return v.GetString().Value(), nil
	case v.IsStruct():
		// convert struct to map
		fields := v.GetStruct().Fields()
		m := make(map[string]any, len(fields))
		for name, val := range fields {
			gv, err := valueToGo(val)
			if err != nil {
				return nil, err
			}
			m[name] = gv
		}
		return m, nil
	case v.IsSymbol():
		return v.GetSymbol().ToProto().Name, nil
	case v.IsTime():
		return v.GetTime().Value(), nil
	}

	return nil, fmt.Errorf("unsupported type: %#v", v)
}

func (py *pySVC) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	py.log.Info("call", zap.String("func", v.String()))
	if py.run.proc == nil {
		py.log.Error("call - python not running")
		return sdktypes.InvalidValue, fmt.Errorf("python not running")
	}

	event := make(map[string]any, len(kwargs))
	for k, v := range kwargs {
		gv, err := valueToGo(v)
		if err != nil {
			py.log.Error("can't convert to Go", zap.Any("value", v), zap.Error(err))
			return sdktypes.InvalidValue, err
		}
		event[k] = gv
	}

	fn := v.GetFunction()
	if !fn.IsValid() {
		py.log.Error("call - invalid function", zap.Any("function", v))
		return sdktypes.InvalidValue, fmt.Errorf("%#v is not a function", v)

	}

	payload, err := json.Marshal(event)
	if err != nil {
		py.log.Error("can't marshal kwargs", zap.Error(err))
		return sdktypes.InvalidValue, err
	}

	fnName := fn.Name().String()
	py.log.Info("call", zap.String("function", fnName))
	if py.firstCall { // TODO: mutex. Ask Itay
		py.firstCall = false

		return py.initialCall(ctx, fnName, payload)
	}

	// Activity call
	msg := PyMessage{
		Type:    "callback",
		Payload: fn.Data(),
	}
	py.log.Info("callback to Python", zap.Any("message", msg))

	if err := py.enc.Encode(msg); err != nil {
		py.log.Error("send to python", zap.Error(err))
		return sdktypes.InvalidValue, err
	}

	var reply PyMessage
	if err := py.dec.Decode(&reply); err != nil {
		py.log.Error("from python", zap.Error(err))
		return sdktypes.InvalidValue, err
	}
	py.log.Info("python return", zap.Any("message", reply))

	return sdktypes.NewBytesValue(reply.Payload), nil
}
