package pythonrt

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
)

type DockerRuntimeConfig struct {
	WorkerAddressProvider func() string
}

type dockerRunnerManager struct {
	logger                *zap.Logger
	client                *dockerClient
	runnerIDToContainerID map[string]string
	mu                    *sync.Mutex
	workerAddressProvider func() string
}

func (c DockerRuntimeConfig) validate() error {
	return nil
}

func ConfigureDockerRunnerManager(log *zap.Logger, cfg DockerRuntimeConfig) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	dc, err := NewDockerClient(log)
	if err != nil {
		return err
	}

	_, err = dc.ensureNetwork()
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("docker connected and synced succesffully, there are %d active runners on network %s", dc.ActiveRunnersCount(), networkName))

	if len(dc.activeRunnerIDs) != 0 {
		log.Info("Stopping orphand runners")
		for rid := range dc.activeRunnerIDs {
			if err := dc.StopRunner(rid); err != nil {
				log.Warn(fmt.Sprintf("failed stopping runner %s: %s", rid, err.Error()))
				continue
			}

			log.Debug(fmt.Sprintf("stopped runner: %s", rid))
		}
	}

	drm := &dockerRunnerManager{
		logger:                log,
		client:                dc,
		runnerIDToContainerID: map[string]string{},
		mu:                    new(sync.Mutex),
		workerAddressProvider: cfg.WorkerAddressProvider,
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
		"workflow",
		"--worker-address",
		workerAddress,
		"--runner-id",
		runnerID,
	}
}

func (rm *dockerRunnerManager) Start(ctx context.Context, buildArtifacts []byte, vars map[string]string) (string, pb.RunnerClient, error) {
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

	if err := rm.client.BuildImage(containerName, codePath); err != nil {
		return "", nil, fmt.Errorf("build image: %w", err)
	}

	runnerID := fmt.Sprintf("runner-%d", time.Now().UnixNano())
	cmd := createStartCommand("main.py", rm.workerAddressProvider(), runnerID)

	cid, port, err := rm.client.StartRunner(containerName, cmd)
	if err != nil {
		return "", nil, fmt.Errorf("start runner: %w", err)
	}

	runnerAddr := fmt.Sprintf("localhost:%s", port)
	client, err := dialRunner(runnerAddr)
	if err != nil {

		if err := rm.client.StopRunner(cid); err != nil {
			rm.logger.Warn("close runner", zap.Error(err))
		}
		return "", nil, err
	}

	rm.mu.Lock()
	rm.runnerIDToContainerID[runnerID] = cid
	rm.mu.Unlock()
	return runnerID, client, nil
}
func (rm *dockerRunnerManager) RunnerHealth(ctx context.Context, runnerID string) error {
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

func (rm *dockerRunnerManager) Stop(ctx context.Context, runnerID string) error {
	rm.mu.Lock()
	cid, ok := rm.runnerIDToContainerID[runnerID]
	rm.mu.Unlock()

	if !ok {
		return errors.New("runner not found")
	}

	return rm.client.StopRunner(cid)
}
func (*dockerRunnerManager) Health(ctx context.Context) error { return nil }
