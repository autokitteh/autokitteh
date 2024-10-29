package pythonrt

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RunnerManager interface {
	Start(ctx context.Context, sessionID sdktypes.SessionID, buildArtifacts []byte, vars map[string]string) (string, *RunnerClient, error)
	RunnerHealth(ctx context.Context, runnerID string) error
	Stop(ctx context.Context, runnerID string) error
	Health(ctx context.Context) error
}

type (
	runnerType   string
	RunnerClient struct {
		pb.UserCodeRunnerServiceClient
		cc *grpc.ClientConn
	}
)

func (c *RunnerClient) Close() error {
	return c.cc.Close()
}

var (
	configuredRunnerType runnerType = runnerTypeNotConfigured
	runnerManager        RunnerManager
)

const (
	runnerTypeLocal         runnerType = "local_runner"
	runnerTypeRemote        runnerType = "remote_runner"
	runnerTypeDocker        runnerType = "docker_runner"
	runnerTypeNotConfigured runnerType = "not_configured"
)

type Healther interface {
	Health(ctx context.Context, in *pb.UserCodeRunnerHealthRequest, opts ...grpc.CallOption) (*pb.UserCodeRunnerHealthResponse, error)
}

func waitForServer(name string, h Healther, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	start := time.Now()
	var req pb.UserCodeRunnerHealthRequest

	for time.Since(start) <= timeout {
		resp, err := h.Health(ctx, &req)
		if err != nil || resp.Error != "" {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		return nil
	}

	return fmt.Errorf("%s not ready after %v", name, timeout)
}

func dialRunner(addr string) (*RunnerClient, error) {
	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	c := RunnerClient{pb.NewUserCodeRunnerServiceClient(conn), conn}

	if err := waitForServer("runner", &c, 10*time.Second); err != nil {
		connCloseErr := conn.Close()
		return nil, errors.Join(err, connCloseErr)
	}
	return &c, nil
}
