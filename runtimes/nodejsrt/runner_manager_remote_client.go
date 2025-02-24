package nodejsrt

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"go.uber.org/zap"

	rmv1 "go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1"
	"go.autokitteh.dev/autokitteh/proto/gen/go/autokitteh/runner_manager/v1/runner_managerv1connect"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type RemoteRuntimeConfig struct {
	ManagerAddress []string
	WorkerAddress  string
}

type remoteRunnerManager struct {
	logger         *zap.Logger
	remoteManagers []runner_managerv1connect.RunnerManagerServiceClient
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
		runner := runner_managerv1connect.NewRunnerManagerServiceClient(http.DefaultClient, addr, connect.WithGRPC())

		resp, err := runner.Health(context.Background(), connect.NewRequest(&rmv1.HealthRequest{}))
		if err != nil {
			return errors.New("could not verify runner manager health")
		}

		if resp.Msg.Error != "" {
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
