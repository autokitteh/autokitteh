package pythonrt

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
	log, err := logger.New(logger.Configs.Dev) // TODO: From configuration
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	info, err := pyExecInfo(ctx)
	if err != nil {
		return nil, err
	}

	log.Info("python info", zap.String("exe", info.Exe), zap.String("version", info.Version))

	svc := pySVC{
		log: log,
	}

	return &svc, nil
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
	Type     string `json:"type"`
	Function string `json:"function"`
	Payload  []byte `json:"payload"`
}

func entriesToValues(xid sdktypes.ExecutorID, entries []string) (map[string]sdktypes.Value, error) {
	values := make(map[string]sdktypes.Value)
	var modFn sdktypes.ModuleFunction // TODO
	for _, name := range entries {
		fn, err := sdktypes.NewFunctionValue(xid, name, nil, nil, modFn)
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

	tarData := compiled[archiveKey]
	if tarData == nil {
		return nil, fmt.Errorf("%q note found in compiled data", archiveKey)
	}

	ri, err := runPython(py.log, tarData, mainPath)
	if err != nil {
		return nil, err
	}

	// Cleanup in case we had errors
	killPy := true
	defer func() {
		if killPy {
			py.log.Error("killing Python", zap.Int("pid", ri.proc.Pid))
			ri.proc.Kill()
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
	/* FIXME: AK calls Close after `Run`, but we need this for `Call` as well.
	if py.run != nil {
		py.run.proc.Kill()
	}
	*/
}

func (py *pySVC) initialCall(ctx context.Context, funcName string, payload []byte) (sdktypes.Value, error) {
	// Initial run cal
	msg := PyMessage{
		Type:     "run",
		Function: funcName,
		Payload:  payload,
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

		var modFn sdktypes.ModuleFunction
		// Generate activity, it'll call Python with the result
		// The function name is irrelevant, all the information Python needs is in the Payload
		fn, err := sdktypes.NewFunctionValue(py.xid, "activity", msg.Payload, nil, modFn)
		if err != nil {
			py.log.Error("create function", zap.Error(err))
			return sdktypes.InvalidValue, err
		}

		py.log.Info("callback")
		py.cbs.Call(
			ctx,
			py.xid.ToRunID(),
			// The Python function to call is encoded in the payload
			fn,
			nil,
			nil,
		)
	}

	py.log.Info("python done, killing")
	py.run.proc.Kill() // FIXME: We run only once
	py.run.proc = nil

	// TODO: Return value
	return sdktypes.Nothing, nil
}

func (py *pySVC) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
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

	fnName := fn.Name().String()
	py.log.Info("call", zap.String("function", fnName))
	if py.firstCall { // TODO: mutex?
		py.firstCall = false
		return py.initialCall(ctx, fnName, fn.Data())
	}

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
