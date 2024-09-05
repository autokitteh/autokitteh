package remotert

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"go.autokitteh.dev/autokitteh/internal/backend/logger"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/runtimes/remotert/pb"

	"go.autokitteh.dev/autokitteh/sdk/sdkruntimes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
)

var (
	Runtime = &sdkruntimes.Runtime{
		Desc: kittehs.Must1(sdktypes.StrictRuntimeFromProto(&sdktypes.RuntimePB{
			Name:           "remote",
			FileExtensions: []string{"py"},
		})),
		New: New,
	}
	config RemoteRuntimeConfig
	runner pb.RunnerManagerClient
)

type svc struct {
	ctx   context.Context // workflow context ?
	runID sdktypes.ExecutorID
	cbs   *sdkservices.RunCallbacks

	firstCall bool
	mainPath  string
	exports   map[string]sdktypes.Value

	// Runner
	runnerID           string
	runnerAddress      string
	runnerClient       pb.RunnerClient
	doneChan           chan []byte // runner report done
	errorChan          chan string // runner report error
	runnerRequestsChan chan *pb.ActivityRequest
}

func (*svc) Get() sdktypes.Runtime {
	return Runtime.Desc
}

func (s *svc) stopRunner(ctx context.Context) {
	_, err := runner.Stop(ctx, &pb.StopRequest{
		RunnerId: s.runnerID,
	})
	if err != nil {
		fmt.Printf("failed stopping runner id %s: %s", s.runnerID, err)
	}
}

func (s *svc) handleInitialCall(ctx context.Context, v sdktypes.FunctionValue, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	defer s.stopRunner(ctx)

	event := make(map[string]any, len(kwargs))
	for key, val := range kwargs {
		goVal, err := unwrap(val)
		if err != nil {
			return sdktypes.InvalidValue, err
		}
		event[key] = goVal
	}

	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return sdktypes.InvalidValue, err
	}

	eventpb := pb.Event{Data: jsonBytes}
	entrypoint := fmt.Sprintf("%s:%s", s.mainPath, v.Name())
	hCtx := metadata.AppendToOutgoingContext(ctx, "runner", s.runnerAddress)
	resp, err := s.runnerClient.Start(hCtx, &pb.StartRequest{EntryPoint: entrypoint, Event: &eventpb})
	if err != nil {
		return sdktypes.InvalidValue, err
	}
	if resp.Error != "" {
		return sdktypes.InvalidValue, errors.New(resp.Error)
	}

	// wait for done
	for {
		select {
		case req := <-s.runnerRequestsChan:
			f := kittehs.Must1(sdktypes.NewFunctionValue(s.runID, req.CallInfo.Function, []byte(req.Data), nil, pyModuleFunc))
			res, err := s.cbs.Call(s.ctx, s.runID.ToRunID(), f, nil, nil)
			response := &pb.ActivityReplyRequest{}
			if err != nil {
				response.Error = err.Error()
			} else {
				response.Data = res.GetBytes().Value()
			}

			hCtx := metadata.AppendToOutgoingContext(ctx, "runner", s.runnerAddress)
			_, err = s.runnerClient.ActivityReply(hCtx, response)
			if err != nil {
				return sdktypes.NewBytesValue([]byte(err.Error())), nil
			}
		case res := <-s.doneChan:
			return sdktypes.NewBytesValue(res), nil
		case err := <-s.errorChan:
			return sdktypes.NewBytesValue([]byte(err)), nil
		}
	}
}

func (s *svc) Call(ctx context.Context, v sdktypes.Value, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	fn := v.GetFunction()

	if s.firstCall {
		s.firstCall = false
		return s.handleInitialCall(ctx, fn, args, kwargs)
	}

	cid := fn.Data()
	resp, err := s.runnerClient.Execute(context.Background(), &pb.ExecuteRequest{Data: cid})

	if err != nil {
		return sdktypes.InvalidValue, err
	}
	if resp.Error != "" {
		return sdktypes.InvalidValue, errors.New(resp.Error)
	}

	return sdktypes.NewBytesValue(resp.Result), nil
}

func (s *svc) ID() sdktypes.RunID {
	return s.runID.ToRunID()
}

func (s *svc) Close() {
	fmt.Println("not closing")
}

func (s *svc) ExecutorID() sdktypes.ExecutorID { return s.runID }

func (s *svc) Values() map[string]sdktypes.Value {
	return s.exports
}

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

// Returns sdktypes.ProgramErrorAsError if not internal error.
func (s *svc) Run(
	ctx context.Context,
	runID sdktypes.RunID, // generated by caller. guaranteed to be unique system-wide.
	mainPath string, // where to start running from.
	compiled map[string][]byte,
	values map[string]sdktypes.Value,
	cbs *sdkservices.RunCallbacks,
) (sdkservices.Run, error) {

	newSvc := s

	newSvc.runID = sdktypes.NewExecutorID(runID) // Should be first
	newSvc.mainPath = mainPath
	newSvc.cbs = cbs
	newSvc.firstCall = true // State for Call.
	newSvc.ctx = ctx

	newSvc.doneChan = make(chan []byte, 1)
	newSvc.errorChan = make(chan string, 1)
	newSvc.runnerRequestsChan = make(chan *pb.ActivityRequest, 1)

	// Load environment defined by user in the `vars` section of the manifest,
	// these are injected to the Python subprocess environment.
	env, err := cbs.Load(ctx, runID, "env")
	if err != nil {
		return nil, fmt.Errorf("can't load env : %w", err)
	}
	envMap := kittehs.TransformMap(env, func(key string, value sdktypes.Value) (string, string) {
		return key, value.GetString().Value()
	})
	// py.log.Info("env", zap.Any("env", envMap))

	tarData := compiled["archive"]
	if tarData == nil {
		return nil, fmt.Errorf("%q note found in compiled data", "archive")
	}

	resp, err := runner.Start(ctx, &pb.StartRunnerRequest{BuildArtifact: tarData, Vars: envMap, WorkerAddress: "0.tcp.eu.ngrok.io:10487"})
	if err != nil {
		return nil, fmt.Errorf("staring runner %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("staring runner %s", resp.Error)
	}

	newSvc.runnerID = resp.RunnerId
	newSvc.runnerAddress = resp.RunnerAddress

	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient("0.0.0.0:7777", grpc.WithTransportCredentials(creds))

	// conn, err := grpc.NewClient(s.runnerAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	newSvc.runnerClient = pb.NewRunnerClient(conn)
	var runnerHealthResponse *pb.HealthResponse
	hCtx := metadata.AppendToOutgoingContext(ctx, "runner", resp.RunnerAddress)
	for {

		runnerHealthResponse, err = newSvc.runnerClient.Health(hCtx, &pb.HealthRequest{})
		if err == nil && runnerHealthResponse.Error == "" {
			break
		}
		fmt.Println("retry health check", newSvc.runnerID)
		time.Sleep(1 * time.Second)
	}

	// ws.svcs[s.runnerID] = s
	ws.mu.Lock()
	ws.runnerIDsToRuntime[newSvc.runnerID] = newSvc
	ws.mu.Unlock()
	// if err != nil {
	// 	return nil, fmt.Errorf("could not verify runner health %w", err)
	// }
	// if runnerHealthResponse.Error != "" {
	// 	return nil, fmt.Errorf("runner health: %w", err)
	// }

	exports, err := newSvc.runnerClient.Exports(hCtx, &pb.ExportsRequest{
		FileName: mainPath,
	})

	if err != nil || exports.Error != "" {
		return nil, fmt.Errorf("failed fetching exports")
	}

	exportsMap, err := entriesToValues(s.runID, exports.Exports)
	if err != nil {
		return nil, err
	}
	newSvc.exports = exportsMap
	// exports.Exports

	return newSvc, nil

}

func Configure(cfg RemoteRuntimeConfig) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	addr := cfg.ManagerAddress[0]

	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}

	runner = pb.NewRunnerManagerClient(conn)
	resp, err := runner.Health(context.Background(), &pb.HealthRequest{})
	if err != nil {
		return fmt.Errorf("could not verify runner manager health")
	}
	if resp.Error != "" {
		return fmt.Errorf("runner manager health: %w", err)
	}

	config = cfg

	return nil
}

func New() (sdkservices.Runtime, error) {
	log, err := logger.New(logger.Configs.Default)
	if err != nil {
		return nil, err
	}

	log = log.With(zap.String("runtime", "remote"))
	log.Debug("init")
	if err := config.validate(); err != nil {
		return nil, fmt.Errorf("remotert: invalid config %w", err)
	}

	if runner == nil {
		return nil, fmt.Errorf("runner not started")
	}

	s := svc{}

	return &s, nil
}
