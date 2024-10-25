package pythonrt

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

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
		pb.RunnerClient
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
	Health(ctx context.Context, in *pb.HealthRequest, opts ...grpc.CallOption) (*pb.HealthResponse, error)
}

func waitForServer(ctx context.Context, sl *zap.SugaredLogger, name string, h Healther, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	tmo := time.After(timeout)

	const retryInterval = 50 * time.Millisecond

	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-tmo:
			return errors.New("timeout waiting for " + name + " server")
		case <-time.After(retryInterval):
			sl.Debugf("polling server health")
			resp, err := h.Health(ctx, &pb.HealthRequest{})
			if err == nil {
				sl.With("resp", resp).Info("got health response after %s: %v", time.Since(startTime), resp)
				if resp.Error != "" {
					return errors.New("error from " + name + " server: " + resp.Error)
				}

				return nil
			}

			sl.With("err", err).Debugf("got error: %v", err)
		}
	}
}

func dialRunner(ctx context.Context, l *zap.Logger, addr string, t time.Duration) (*RunnerClient, error) {
	sl := l.Sugar().With("addr", addr)

	sl.Infof("dialing runner at %q", addr)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	c := RunnerClient{pb.NewRunnerClient(conn), conn}

	if err := waitForServer(ctx, sl, "runner", &c, t); err != nil {
		connCloseErr := conn.Close()
		return nil, errors.Join(err, connCloseErr)
	}

	return &c, nil
}
