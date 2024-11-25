package pythonrt

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"go.autokitteh.dev/autokitteh/internal/backend/tar"
	rmv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1/runner_managerv1connect"
	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/user_code/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type RemoteRuntimeConfig struct {
	ManagerAddress []string
	WorkerAddress  string
}

type managedRunnerManagerClient struct {
	client runner_managerv1connect.RunnerManagerServiceClient
	addr   string
}

func newManagedRunnerManager(addr string) (*managedRunnerManagerClient, error) {
	runner := runner_managerv1connect.NewRunnerManagerServiceClient(http.DefaultClient, addr, connect.WithGRPC())

	resp, err := runner.Health(context.Background(), connect.NewRequest(&rmv1.HealthRequest{}))
	if err != nil {
		return nil, fmt.Errorf("could not verify runner manager health")
	}

	if resp.Msg.Error != "" {
		return nil, fmt.Errorf("runner manager health: %w", err)
	}

	return &managedRunnerManagerClient{
		client: runner,
		addr:   addr,
	}, nil
}

type remoteRunnersManager struct {
	managedRunnerManagers []*managedRunnerManagerClient
	managersLock          *sync.Mutex
	nextClientIndex       int
	activeUseCount        map[int]int
	maxUsePerClient       int
	logger                *zap.Logger

	runnersLock             *sync.Mutex
	runnerIDToRemoteManager map[string]*managedRunnerManagerClient //map runnerID to runner manager
	runnerIDToClient        map[string]*RunnerClient               //map runnerID to RunnerClient

	cfg RemoteRuntimeConfig
}

func newRemoteRunnerManager(l *zap.Logger, cfg RemoteRuntimeConfig) *remoteRunnersManager {
	return &remoteRunnersManager{
		managedRunnerManagers:   make([]*managedRunnerManagerClient, 0),
		managersLock:            &sync.Mutex{},
		runnersLock:             &sync.Mutex{},
		runnerIDToRemoteManager: make(map[string]*managedRunnerManagerClient),
		runnerIDToClient:        make(map[string]*RunnerClient),
		nextClientIndex:         0,
		activeUseCount:          make(map[int]int),
		maxUsePerClient:         50, //TODO: get this from the client, need to extened rpc
		logger:                  l,
		cfg:                     cfg,
	}
}

func (p *remoteRunnersManager) addRemoteManager(cli *managedRunnerManagerClient) {
	p.managersLock.Lock()
	defer p.managersLock.Unlock()
	p.managedRunnerManagers = append(p.managedRunnerManagers, cli)
}

// func (p *remoteRunnersManager) removeRemoteManager(cli *managedRunnerManagerClient) {
// 	p.managersLock.Lock()
// 	defer p.managersLock.Unlock()
// 	for i, c := range p.managedRunnerManagers {
// 		if c == cli {
// 			p.managedRunnerManagers = append(p.managedRunnerManagers[:i], p.managedRunnerManagers[i+1:]...)
// 			return
// 		}
// 	}
// }

func (p *remoteRunnersManager) acquireRemoteManager() (*managedRunnerManagerClient, error) {
	p.managersLock.Lock()
	defer p.managersLock.Unlock()
	if len(p.managedRunnerManagers) == 0 {
		return nil, errors.New("no clients available")
	}

	// Algorithm, find the client with the least active use count
	var cli *managedRunnerManagerClient
	minUsage := p.maxUsePerClient

	for i, c := range p.managedRunnerManagers {
		if p.activeUseCount[i] < minUsage {
			cli = c
			minUsage = p.activeUseCount[i]
		}
	}

	if cli == nil {
		return nil, errors.New("no clients available")
	}
	return cli, nil
}

func (p *remoteRunnersManager) releaseRemoteManager(cli *managedRunnerManagerClient) {
	p.managersLock.Lock()
	defer p.managersLock.Unlock()

	for i, c := range p.managedRunnerManagers {
		if c == cli {
			p.activeUseCount[i]--
			return
		}
	}

}

func (c RemoteRuntimeConfig) validate() error {
	if len(c.ManagerAddress) == 0 {
		return errors.New("no runner manager address")
	}

	return nil
}

func configureRemoteRunnerManager(log *zap.Logger, cfg RemoteRuntimeConfig) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	if configuredRunnerType != runnerTypeNotConfigured {
		return errors.New("runner type already configured, cannot configure twice")
	}

	log.Info("remote runner worker address", zap.String("worker_address", cfg.WorkerAddress))
	rrm := newRemoteRunnerManager(log, cfg)

	for _, addr := range cfg.ManagerAddress {
		cli, err := newManagedRunnerManager(addr)
		if err != nil {
			return err
		}
		rrm.addRemoteManager(cli)

	}

	configuredRunnerType = runnerTypeRemote
	runnerManager = rrm
	rrm.logger.Info("configured")
	return nil
}

func createRunnerAuthInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		mdCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs("token", token))
		return invoker(mdCtx, method, req, reply, cc, opts...)
	}
}

func (rrm *remoteRunnersManager) Start(ctx context.Context, sessionID sdktypes.SessionID, buildArtifacts []byte, vars map[string]string) (string, *RunnerClient, error) {
	runnerManager, err := rrm.acquireRemoteManager()
	if err != nil {
		return "", nil, err
	}

	rrm.logger.Info(fmt.Sprintf("using runner manager %s", runnerManager.addr))

	userCodePath, err := prepareUserCode(buildArtifacts, false)
	if err != nil {
		return "", nil, err
	}

	t := tar.NewTarFile()

	filepath.Walk(userCodePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		filePath := strings.TrimPrefix(path, userCodePath+"/")
		t.Add(filePath, data)
		return nil
	})

	buildBytes, err := t.Bytes(true)
	if err != nil {
		return "", nil, err
	}

	resp, err := runnerManager.client.StartRunner(ctx, connect.NewRequest(&rmv1.StartRunnerRequest{
		SessionId: sessionID.String(),
		RunnerConfig: &rmv1.StartRunnerRequest_UserCode{
			UserCode: &rmv1.UserCode{
				BuildArtifact: buildBytes,
			},
		},
		Vars:          vars,
		WorkerAddress: rrm.cfg.WorkerAddress,
	}))
	if err != nil {
		return "", nil, err
	}

	if resp.Msg.Error != "" {
		return "", nil, fmt.Errorf("start runner: %w", err)
	}

	client, err := dialRunner(ctx, resp.Msg.RunnerAddress, createRunnerAuthInterceptor(resp.Msg.RunnerToken))
	if err != nil {
		stopResp, stopError := runnerManager.client.StopRunner(ctx, connect.NewRequest(&rmv1.StopRunnerRequest{RunnerId: resp.Msg.RunnerId}))
		if stopError != nil {
			rrm.logger.Warn(fmt.Sprintf("close runner %s: %s", resp.Msg.RunnerId, stopError))
		}

		if stopResp.Msg.Error != "" {
			rrm.logger.Warn(fmt.Sprintf("close runner %s: %s", resp.Msg.RunnerId, resp.Msg.Error))
		}
		return "", nil, err
	}

	return resp.Msg.RunnerId, client, nil
}
func (rrm *remoteRunnersManager) RunnerHealth(ctx context.Context, runnerID string) error {
	rrm.runnersLock.Lock()
	defer rrm.runnersLock.Unlock()
	cli, ok := rrm.runnerIDToClient[runnerID]
	if !ok {
		return errors.New("runner not found")
	}
	req := &pb.RunnerHealthRequest{}
	resp, err := cli.Health(ctx, req)
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.New(resp.Error)
	}
	return nil
}

func (rrm *remoteRunnersManager) Stop(ctx context.Context, runnerID string) error {
	rrm.runnersLock.Lock()
	defer rrm.runnersLock.Unlock()
	cli, ok := rrm.runnerIDToClient[runnerID]
	if !ok {
		return errors.New("runner not found")
	}
	if err := cli.Close(); err != nil {
		rrm.logger.Warn(fmt.Sprintf("close runner %s failed %s", runnerID, err))
	}

	manager, ok := rrm.runnerIDToRemoteManager[runnerID]
	if !ok {
		return errors.New("runner not found")
	}

	rrm.releaseRemoteManager(manager)

	return nil
}
func (*remoteRunnersManager) Health(ctx context.Context) error { return nil }
