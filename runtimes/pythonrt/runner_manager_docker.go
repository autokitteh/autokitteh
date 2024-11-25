package pythonrt

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"sync"

	"go.jetify.com/typeid"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type DockerRuntimeConfig struct {
	WorkerAddressProvider func() string
	LogRunnerCode         bool
	LogBuildCode          bool
}

type DockerRunnerManager struct {
	logger                *zap.Logger
	client                *DockerClient
	runnerIDToContainerID map[string]string
	mu                    *sync.Mutex
	workerAddressProvider func() string
}

func NewDockerRunnerManager(log *zap.Logger, cfg DockerRuntimeConfig) (*DockerRunnerManager, error) {
	dc, err := NewDockerClient(log, cfg.LogRunnerCode, cfg.LogBuildCode)
	if err != nil {
		return nil, err
	}

	_, err = dc.EnsureNetwork()
	if err != nil {
		return nil, err
	}

	log.Info(fmt.Sprintf("docker connected and synced succesffully, there are %d active runners on network %s", dc.ActiveRunnersCount(), networkName))

	// we don't reconnect to existing runners, we start new ones
	// so in case server started and there are some runners running
	// we stop them
	dc.CleanupOnStart()

	drm := &DockerRunnerManager{
		logger:                log,
		client:                dc,
		runnerIDToContainerID: map[string]string{},
		mu:                    new(sync.Mutex),
		workerAddressProvider: cfg.WorkerAddressProvider,
	}
	return drm, nil
}

func configureDockerRunnerManager(log *zap.Logger, cfg DockerRuntimeConfig) error {
	drm, err := NewDockerRunnerManager(log, cfg)
	if err != nil {
		return err
	}

	configuredRunnerType = runnerTypeDocker
	runnerManager = drm
	drm.logger.Info("configured")
	return nil
}

func createStartCommand(entrypoint, workerAddress, runnerID string) []string {
	// ["main.py", "--code-dir", "workflow", "--worker-address", "host.docker.internal:9980"]
	return []string{
		entrypoint,
		"--code-dir",
		"/workflow",
		"--worker-address",
		workerAddress,
		"--runner-id",
		runnerID,
	}
}

func (rm *DockerRunnerManager) Start(ctx context.Context, sessionID sdktypes.SessionID, buildArtifacts []byte, vars map[string]string) (string, *RunnerClient, error) {
	if len(buildArtifacts) == 0 {
		return "", nil, errors.New("no build artifacts")
	}

	codePath, err := prepareUserCode(buildArtifacts, false)
	if err != nil {
		return "", nil, fmt.Errorf("prepare user code: %w", err)
	}

	hash := md5.Sum(buildArtifacts)
	version := fmt.Sprintf("u%x", hash)
	containerName := fmt.Sprintf("usercode:%s", version)

	if err := rm.client.BuildImage(ctx, containerName, codePath); err != nil {
		return "", nil, fmt.Errorf("build image: %w", err)
	}

	rid, err := typeid.WithPrefix("runner")
	if err != nil {
		return "", nil, err
	}
	runnerID := rid.String()
	cmd := createStartCommand("/runner/main.py", rm.workerAddressProvider(), runnerID)

	cid, port, err := rm.client.StartRunner(ctx, containerName, sessionID, cmd, vars)
	if err != nil {
		return "", nil, fmt.Errorf("start runner: %w", err)
	}

	runnerAddr := fmt.Sprintf("127.0.0.1:%s", port)
	client, err := dialRunner(ctx, runnerAddr)
	if err != nil {

		if err := rm.client.StopRunner(ctx, cid); err != nil {
			rm.logger.Warn("close runner", zap.Error(err))
		}
		return "", nil, err
	}

	rm.mu.Lock()
	rm.runnerIDToContainerID[runnerID] = cid
	rm.mu.Unlock()
	return runnerID, client, nil
}
func (rm *DockerRunnerManager) RunnerHealth(ctx context.Context, runnerID string) error {
	rm.mu.Lock()
	cid, ok := rm.runnerIDToContainerID[runnerID]
	rm.mu.Unlock()

	if !ok {
		return errors.New("runner not found")
	}

	isRunning, err := rm.client.IsRunning(cid)
	if err != nil {
		return err
	}
	if !isRunning {
		return errors.New("runner not running")
	}

	return nil
}

func (rm *DockerRunnerManager) Stop(ctx context.Context, runnerID string) error {
	rm.mu.Lock()
	cid, ok := rm.runnerIDToContainerID[runnerID]
	rm.mu.Unlock()

	if !ok {
		return errors.New("runner not found")
	}

	return rm.client.StopRunner(ctx, cid)
}
func (*DockerRunnerManager) Health(ctx context.Context) error { return nil }
