package pythonrt

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap"
	"go.uber.org/zap/zapio"
)

const (
	runnersLabel       = "io.autokitteh.cloud.runner"
	networkName        = "autokitteh_runners"
	internalRunnerPort = "9293/tcp"
)

type dockerClient struct {
	client          *client.Client
	activeRunnerIDs map[string]struct{}
	allRunnerIDs    map[string]struct{}
	mu              *sync.Mutex
	runnerLabels    map[string]string
	logBuildProcess bool
	logRunner       bool
	logger          *zap.Logger
	cpuNano         int64
	memoryMB        int64
}

func NewDockerClient(logger *zap.Logger, logRunner, logBuildProcess bool, cpuPerContainer float32, memoryPerContainer uint32) (*dockerClient, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	if cpuPerContainer < 0.1 {
		return nil, errors.New("cpu per container must be at least 0.1")
	}

	if memoryPerContainer < 64 {
		return nil, errors.New("memory per container must be at least 64 MB")
	}

	cpuNano := int64(cpuPerContainer * 100_000_000)
	memoryMB := int64(memoryPerContainer) * 1024 * 1024

	dc := &dockerClient{
		client:          apiClient,
		mu:              new(sync.Mutex),
		runnerLabels:    map[string]string{runnersLabel: ""},
		activeRunnerIDs: map[string]struct{}{},
		allRunnerIDs:    map[string]struct{}{},
		logger:          logger,
		logBuildProcess: logBuildProcess,
		logRunner:       logRunner,
		cpuNano:         cpuNano, // 0.5 CPUs (in nanoseconds)
		memoryMB:        memoryMB,
	}

	if err := dc.SyncCurrentState(); err != nil {
		return nil, err
	}

	return dc, nil
}

func (d *dockerClient) ensureNetwork() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inspectResult, err := d.client.NetworkInspect(ctx, networkName, network.InspectOptions{})
	if err != nil {
		if !client.IsErrNotFound(err) {
			return "", err
		}

		noICCNetworkOptions := map[string]string{"com.docker.network.bridge.enable_icc": "false"}
		n, err := d.client.NetworkCreate(ctx, networkName, network.CreateOptions{Options: noICCNetworkOptions})
		if err != nil {
			return "", err
		}
		return n.ID, nil
	}

	noICCOption := inspectResult.Options["com.docker.network.bridge.enable_icc"]
	if noICCOption != "false" {
		return "", errors.New("network with invalid icc, need to recreate")
	}

	return inspectResult.ID, nil
}

func (d *dockerClient) StartRunner(ctx context.Context, runnerImage string, sessionID sdktypes.SessionID, cmd []string, vars map[string]string) (string, string, error) {
	envVars := make([]string, 0, len(vars))
	for k, v := range vars {
		envVars = append(envVars, k+"="+v)
	}

	resources := container.Resources{
		NanoCPUs: d.cpuNano,  // 0.5 CPUs (in nanoseconds)
		Memory:   d.memoryMB, // 128 MB
	}
	resp, err := d.client.ContainerCreate(ctx,
		&container.Config{
			Image: runnerImage,
			Cmd:   cmd,
			Tty:   false,
			Env:   envVars,
			ExposedPorts: map[nat.Port]struct{}{
				nat.Port(internalRunnerPort): {},
			},
			Labels:     d.runnerLabels,
			WorkingDir: "/workflow",
		},
		&container.HostConfig{
			NetworkMode:    container.NetworkMode(networkName),
			PortBindings:   nat.PortMap{internalRunnerPort: []nat.PortBinding{{HostIP: "0.0.0.0"}}},
			Tmpfs:          map[string]string{"/tmp": "size=64m"},
			Resources:      resources,
			ReadonlyRootfs: true,
		}, nil, nil, "")

	if err != nil {
		return "", "", err
	}

	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", "", err
	}

	d.mu.Lock()
	d.allRunnerIDs[resp.ID] = struct{}{}
	d.mu.Unlock()

	d.setupContainerLogging(ctx, resp.ID, sessionID)
	port, err := d.nextFreePort(ctx, resp.ID)
	if err != nil {
		return "", "", err
	}

	d.mu.Lock()
	d.activeRunnerIDs[resp.ID] = struct{}{}
	d.mu.Unlock()

	return resp.ID, port, nil
}

func (d *dockerClient) nextFreePort(ctx context.Context, cid string) (string, error) {

	for i := 0; i < 10; i++ {
		inspect, err := d.client.ContainerInspect(ctx, cid)
		if err != nil {
			return "", err
		}

		ports, ok := inspect.NetworkSettings.Ports[nat.Port(internalRunnerPort)]
		if ok && len(ports) > 0 {
			port := ports[0].HostPort
			return port, nil
		}

		time.Sleep(time.Second)
	}

	return "", errors.New("couldn't find port")
}

func (d *dockerClient) setupContainerLogging(ctx context.Context, cid string, sessionID sdktypes.SessionID) {
	go func() {
		reader, _ := d.client.ContainerLogs(ctx, cid, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		defer reader.Close()
		var err error
		l := d.logger.With(zap.String("container_id", cid), zap.String("session_id", sessionID.String()))

		if d.logRunner {
			stdourWriter := zapio.Writer{Log: l.With(zap.String("stream", "stdout"))}
			stderrWriter := zapio.Writer{Log: l.With(zap.String("stream", "stderr"))}
			_, err = stdcopy.StdCopy(&stdourWriter, &stderrWriter, reader)
			defer stdourWriter.Close()
			defer stderrWriter.Close()
		} else {
			_, _ = io.Copy(io.Discard, reader)
		}

		if err != nil {
			l.Warn("error reading container logs", zap.Error(err))
		}
	}()
}

func (d *dockerClient) SyncCurrentState() error {
	listedContainers, err := d.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// reset the state
	d.allRunnerIDs = map[string]struct{}{}
	d.activeRunnerIDs = map[string]struct{}{}

	for _, c := range listedContainers {
		if _, ok := c.Labels[runnersLabel]; !ok {
			continue
		}
		d.allRunnerIDs[c.ID] = struct{}{}

		if c.State != "running" {
			continue
		}

		d.activeRunnerIDs[c.ID] = struct{}{}
	}

	return nil
}

func (d *dockerClient) ImageExists(ctx context.Context, imageName string) (bool, error) {
	images, err := d.client.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return false, err
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				return true, nil
			}
		}
	}

	return false, nil
}

func (d *dockerClient) BuildImage(ctx context.Context, name, directory string) error {
	tar, err := archive.TarWithOptions(directory, &archive.TarOptions{})
	if err != nil {
		return err
	}

	options := types.ImageBuildOptions{
		Dockerfile: "Dockerfile", // Name of the Dockerfile
		Tags:       []string{name},
		Remove:     true,
	}

	// Build the image
	resp, err := d.client.ImageBuild(ctx, tar, options)
	if err != nil {
		d.logger.Error("Error building image", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	var dest io.Writer = io.Discard
	if d.logBuildProcess {
		dest = os.Stdout
	}
	if _, err := io.Copy(dest, resp.Body); err != nil {
		d.logger.Error("Error printing build output", zap.Error(err))
		return err
	}

	exists, err := d.ImageExists(ctx, name)
	if err != nil {
		return err
	}

	if !exists {
		return errors.New("failed creating image")
	}

	d.logger.Info("Image built successfully")
	return nil
}

func (d *dockerClient) ActiveRunnersCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.activeRunnerIDs)
}

func (d *dockerClient) GetActiveRunners() map[string]struct{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.activeRunnerIDs
}

func (d *dockerClient) IsRunning(runnerID string) (bool, error) {
	if err := d.SyncCurrentState(); err != nil {
		return false, err
	}

	_, ok := d.activeRunnerIDs[runnerID]
	return ok, nil

}

func (d *dockerClient) StopRunner(ctx context.Context, id string) error {
	// this is to unlock as fast as possible
	// since stopping a container can take a while
	d.mu.Lock()
	_, isRunnerID := d.allRunnerIDs[id]
	if isRunnerID {
		delete(d.allRunnerIDs, id)
	}
	_, isActiveRunner := d.activeRunnerIDs[id]
	if isActiveRunner {
		delete(d.activeRunnerIDs, id)
	}
	d.mu.Unlock()

	if !isRunnerID {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if isActiveRunner {
		timeout := 0 // default to 0, kill now
		err := d.client.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
		if err != nil {
			return err
		}
	}

	if err := d.client.ContainerRemove(ctx, id, container.RemoveOptions{}); err != nil {
		return err
	}

	return nil
}
