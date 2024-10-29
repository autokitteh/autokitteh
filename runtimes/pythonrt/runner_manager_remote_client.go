package pythonrt

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/remote/v1"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type RemoteRuntimeConfig struct {
	ManagerAddress []string
	WorkerAddress  string
}

type remoteRunnerManager struct {
	logger         *zap.Logger
	remoteManagers []pb.RunnerManagerServiceClient
}

func (c RemoteRuntimeConfig) validate() error {
	if len(c.ManagerAddress) == 0 {
		return errors.New("no runner manager address")
	}

	return nil
}

func configureRemoteRunnerManager(cfg RemoteRuntimeConfig) error {
	if err := cfg.validate(); err != nil {
		return err
	}

	if configuredRunnerType != runnerTypeNotConfigured {
		return errors.New("runner type already configured, cannot configure twice")
	}

	rrm := &remoteRunnerManager{}

	for _, addr := range cfg.ManagerAddress {
		creds := insecure.NewCredentials()
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
		if err != nil {
			return err
		}

		runner := pb.NewRunnerManagerServiceClient(conn)
		resp, err := runner.Health(context.Background(), &pb.RunnerManagerHealthRequest{})
		if err != nil {
			return fmt.Errorf("could not verify runner manager health")
		}
		if resp.Error != "" {
			return fmt.Errorf("runner manager health: %w", err)
		}
		rrm.remoteManagers = append(rrm.remoteManagers, runner)
	}

	configuredRunnerType = runnerTypeRemote
	runnerManager = rrm
	rrm.logger.Info("configured")
	return nil
}

func (rrm *remoteRunnerManager) Start(ctx context.Context, sessionID sdktypes.SessionID, buildArtifacts []byte, vars map[string]string) (string, *RunnerClient, error) {
	// rrm.remoteManagers[0].StartRunner(ctx, &pb.StartRunnerRequest{})
	return "", nil, nil
}
func (*remoteRunnerManager) RunnerHealth(ctx context.Context, runnerID string) error { return nil }
func (*remoteRunnerManager) Stop(ctx context.Context, runnerID string) error         { return nil }
func (*remoteRunnerManager) Health(ctx context.Context) error                        { return nil }
