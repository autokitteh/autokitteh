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
}

func NewDockerClient(logger *zap.Logger, logRunner, logBuildProcess bool) (*dockerClient, error) {
	apiClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	dc := &dockerClient{
		client:          apiClient,
		mu:              new(sync.Mutex),
		runnerLabels:    map[string]string{runnersLabel: ""},
		activeRunnerIDs: map[string]struct{}{},
		allRunnerIDs:    map[string]struct{}{},
		logger:          logger,
		logBuildProcess: logBuildProcess,
		logRunner:       logRunner,
	}

	if err := dc.SyncCurrentState(); err != nil {
		return nil, err
	}

	return dc, nil
}

// func (d *dockerClient) close() error {
// 	return d.client.Close()
// }

func (d *dockerClient) ensureNetwork() (string, error) {
	inspectResult, err := d.client.NetworkInspect(context.Background(), networkName, network.InspectOptions{})
	if err != nil {
		if !client.IsErrNotFound(err) {
			return "", err
		}

		noICCNetworkOptions := map[string]string{"com.docker.network.bridge.enable_icc": "false"}
		n, err := d.client.NetworkCreate(context.Background(), networkName, network.CreateOptions{Options: noICCNetworkOptions})
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

func (d *dockerClient) StartRunner(runnerImage string, sessionID sdktypes.SessionID, cmd []string) (string, string, error) {
	resp, err := d.client.ContainerCreate(context.Background(),
		&container.Config{
			Image: runnerImage,
			Cmd:   cmd,
			Tty:   false,

			ExposedPorts: map[nat.Port]struct{}{
				nat.Port(internalRunnerPort): {},
			},
			Labels: d.runnerLabels,
		},
		&container.HostConfig{
			NetworkMode:  container.NetworkMode(networkName),
			PortBindings: nat.PortMap{internalRunnerPort: []nat.PortBinding{{HostIP: "0.0.0.0"}}},
		}, nil, nil, "")

	if err != nil {
		return "", "", err
	}

	if err := d.client.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		return "", "", err
	}

	d.mu.Lock()
	d.allRunnerIDs[resp.ID] = struct{}{}
	d.mu.Unlock()

	go func() {
		reader, _ := d.client.ContainerLogs(context.Background(), resp.ID, container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		defer reader.Close()

		if d.logRunner {
			l := d.logger.With(zap.String("runner_id", resp.ID), zap.String("session_id", sessionID.String()))
			stdourWriter := zapio.Writer{Log: l.With(zap.String("stream", "stdout"))}
			stderrWriter := zapio.Writer{Log: l.With(zap.String("stream", "stderr"))}
			_, err = stdcopy.StdCopy(&stdourWriter, &stderrWriter, reader)
			defer stdourWriter.Close()
			defer stderrWriter.Close()
		} else {
			_, err = io.Copy(io.Discard, reader)
		}

		if err != nil {
			d.logger.Warn("error reading runner logs", zap.Error(err), zap.String("runner_id", resp.ID))
		}
	}()

	var port string

	for i := 0; i < 10; i++ {
		inspect, err := d.client.ContainerInspect(context.Background(), resp.ID)
		if err != nil {
			return "", "", err
		}

		ports, ok := inspect.NetworkSettings.Ports[nat.Port(internalRunnerPort)]
		if ok && len(ports) > 0 {
			port = ports[0].HostPort
			break
		}

		time.Sleep(1000 * time.Millisecond)
	}

	if port == "" {
		return "", "", errors.New("couldn't find port")
	}

	d.mu.Lock()
	d.activeRunnerIDs[resp.ID] = struct{}{}
	d.mu.Unlock()

	return resp.ID, port, nil
}

func (d *dockerClient) SyncCurrentState() error {
	listedContainers, err := d.client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return err
	}

	d.mu.Lock()
	defer d.mu.Unlock()

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

func (d *dockerClient) ImageExists(imageName string) (bool, error) {
	images, err := d.client.ImageList(context.Background(), image.ListOptions{All: true})
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

func (d *dockerClient) BuildImage(name, directory string) error {
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
	resp, err := d.client.ImageBuild(context.Background(), tar, options)
	if err != nil {
		d.logger.Error("Error building image", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	if d.logBuildProcess {
		_, err = io.Copy(os.Stdout, resp.Body)
	} else {
		_, err = io.Copy(io.Discard, resp.Body)
	}

	if err != nil {
		d.logger.Error("Error printing build output", zap.Error(err))
		return err
	}

	exists, err := d.ImageExists(name)
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
	return len(d.activeRunnerIDs)
}

func (d *dockerClient) GetActiveRunners() map[string]struct{} {
	return d.activeRunnerIDs
}

func (d *dockerClient) IsRunning(runnerID string) (bool, error) {
	if err := d.SyncCurrentState(); err != nil {
		return false, err
	}

	_, ok := d.activeRunnerIDs[runnerID]
	return ok, nil

}

func (d *dockerClient) StopRunner(id string) error {
	if _, ok := d.allRunnerIDs[id]; !ok {
		return nil
	}

	if _, ok := d.activeRunnerIDs[id]; ok {
		var timeout int // default to 0, kill now
		err := d.client.ContainerStop(context.Background(), id, container.StopOptions{Timeout: &timeout})
		if err != nil {
			return err
		}
	}

	if err := d.client.ContainerRemove(context.Background(), id, container.RemoveOptions{}); err != nil {
		return err
	}

	delete(d.allRunnerIDs, id)
	delete(d.activeRunnerIDs, id)

	return nil
}
