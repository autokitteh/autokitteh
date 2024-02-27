package pythonrt

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"time"

	"go.autokitteh.dev/autokitteh/backend/logger"
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
	values    map[string]sdktypes.Value
	firstCall bool
	dec       *json.Decoder
	enc       *json.Encoder
}

func New() (sdkservices.Runtime, error) {
	log, err := logger.New(logger.Configs.Dev) // TODO: From configuration
	if err != nil {
		return nil, err
	}
	log = log.With(zap.String("runtime", "python"))

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
		py.log.Error("build", zap.Error(err))
		return nil, err
	}

	art, err := sdktypes.BuildArtifactFromProto(&sdktypes.BuildArtifactPB{
		CompiledData: map[string][]byte{
			archiveKey: data,
		},
	})

	if err != nil {
		return nil, err
	}

	return art, nil
}

type PyMessage struct {
	Type     string `json:"type"`
	Function string `json:"function"`
	Payload  []byte `json:"payload"`
}

func entriesToValues(xid sdktypes.ExecutorID, entries []string) map[string]sdktypes.Value {
	values := make(map[string]sdktypes.Value)
	for _, name := range entries {
		values[name] = sdktypes.NewFunctionValue(xid, name, nil, nil, nil)
	}

	return values
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
	tarData := compiled[archiveKey]
	if tarData == nil {
		return nil, fmt.Errorf("%q note found in compiled data", archiveKey)
	}

	ri, err := runPython(py.log, tarData, mainPath)
	if err != nil {
		return nil, err
	}

	killPy := true
	defer func() {
		if killPy {
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

	var entries []string
	if err := json.Unmarshal(msg.Payload, &entries); err != nil {
		py.log.Error("can't parse module entries", zap.Error(err))
		return nil, fmt.Errorf("can't parse module entries: %w", err)
	}

	killPy = false
	py.run = ri
	py.cbs = cbs
	py.xid = sdktypes.NewExecutorID(runID)
	py.values = entriesToValues(py.xid, entries)
	py.firstCall = true
	py.dec = dec
	py.enc = json.NewEncoder(conn)

	return py, nil
}

func (py *pySVC) ID() sdktypes.RunID              { return py.xid.ToRunID() }
func (py *pySVC) ExecutorID() sdktypes.ExecutorID { return py.xid }

func (py *pySVC) Values() map[string]sdktypes.Value {
	return py.values
}

func (py *pySVC) Close() {
	if py.run != nil {
		py.run.proc.Kill()
	}
}

func (py *pySVC) initialCall(ctx context.Context, funcName string, payload []byte) (sdktypes.Value, error) {
	// Initial run cal
	msg := PyMessage{
		Type:     "run",
		Function: funcName,
		Payload:  payload,
	}
	py.log.Info("run", zap.Any("message", msg))
	if err := py.enc.Encode(msg); err != nil {
		return nil, err
	}

	// Callbacks
	for {
		var msg PyMessage
		if err := py.dec.Decode(&msg); err != nil {
			py.log.Error("communication error", zap.Error(err))
			return nil, err
		}
		py.log.Info("from python", zap.Any("message", msg))

		if msg.Type == "done" {
			break
		}

		// Generate activity, it'll call Python with the result
		py.cbs.Call(
			ctx,
			py.xid.ToRunID(),
			// The function to call is encoded in the payload
			sdktypes.NewFunctionValue(py.xid, "activity", msg.Payload, nil, nil),
			nil,
			nil,
		)
	}

	// TODO: Return value
	return nil, nil
}

func (py *pySVC) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	fn := sdktypes.GetFunctionValue(v).ToProto()
	if py.firstCall { // TODO: mutex?
		py.firstCall = false
		return py.initialCall(ctx, fn.Name, fn.Data)
	}

	msg := PyMessage{
		Type:    "callback",
		Payload: fn.Data,
	}

	if err := py.enc.Encode(msg); err != nil {
		py.log.Error("send to python", zap.Error(err))
		return nil, err
	}

	var reply PyMessage
	if err := py.dec.Decode(&reply); err != nil {
		py.log.Error("from python", zap.Error(err))
		return nil, err
	}

	return sdktypes.NewBytesValue(reply.Payload), nil
}
