package pythonrt

import (
	"context"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
	userCode "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/user_code/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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
		userCode.RunnerServiceClient
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

type Healthier interface {
	Health(ctx context.Context, in *userCode.RunnerHealthRequest, opts ...grpc.CallOption) (*userCode.RunnerHealthResponse, error)
}

func dialRunner(ctx context.Context, addr string) (*RunnerClient, error) {
	ctx, span := telemetry.T().Start(ctx, "dialRunner")
	defer span.End()

	span.SetAttributes(attribute.String("addr", addr))

	creds := insecure.NewCredentials()
	// Python takes it's time going up, this with grpc.WaitForReady(true) below
	// makes the connection wait until Python is ready.
	params := grpc.ConnectParams{
		MinConnectTimeout: 10 * time.Second,
	}
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithConnectParams(params),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithTracerProvider(telemetry.TraceProvider()),
			otelgrpc.WithMeterProvider(telemetry.MetricProvider()),
		)),
	)
	if err != nil {
		return nil, err
	}

	c := RunnerClient{userCode.NewRunnerServiceClient(conn), conn}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := c.Health(ctx, &userCode.RunnerHealthRequest{}, grpc.WaitForReady(true)); err != nil {
		return nil, err
	}

	return &c, nil
}
