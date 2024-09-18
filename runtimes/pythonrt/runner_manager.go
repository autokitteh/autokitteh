package pythonrt

import (
	"context"
	"fmt"
	"time"

	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RunnerManager interface {
	Start(ctx context.Context, buildArtifacts []byte, vars map[string]string) (string, pb.RunnerClient, error)
	RunnerHealth(ctx context.Context, runnerID string) error
	Stop(ctx context.Context, runnerID string) error
	Health(ctx context.Context) error
}

type runnerType string

var (
	configuredRunnerType runnerType = runnerTypeNotConfigured
	runnerManager        RunnerManager
)

const (
	runnerTypeLocal         runnerType = "local_runner"
	runnerTypeRemote        runnerType = "remote_runner"
	runnerTypeNotConfigured runnerType = "not_configured"
)

type Healther interface {
	Health(ctx context.Context, in *pb.HealthRequest, opts ...grpc.CallOption) (*pb.HealthResponse, error)
}

func waitForServer(name string, h Healther, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	start := time.Now()
	var req pb.HealthRequest

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

func dialRunner(addr string) (pb.RunnerClient, error) {
	creds := insecure.NewCredentials()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, err
	}

	c := pb.NewRunnerClient(conn)

	if err := waitForServer("runner", c, 10*time.Second); err != nil {
		return nil, err
	}
	return c, nil
}
