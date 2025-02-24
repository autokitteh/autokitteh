package nodejsrt

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"sync"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type localRunnerManager struct {
	logger           *zap.Logger
	nodeExe          string
	runnerIDToRunner map[string]*LocalNodeJS
	mu               *sync.Mutex
	workerAddress    string
	cfg              LocalRunnerManagerConfig
}

type LocalRunnerManagerConfig struct {
	WorkerAddress         string
	LazyLoadVEnv          bool
	WorkerAddressProvider func() string
	LogCodeRunnerCode     bool
}

func configureLocalRunnerManager(log *zap.Logger, cfg LocalRunnerManagerConfig) error {
	if cfg.WorkerAddress == "" && cfg.WorkerAddressProvider == nil {
		return errors.New("either workerAddress or workerAddressProvider should be supplied")
	}

	lm := &localRunnerManager{
		logger:           log,
		runnerIDToRunner: map[string]*LocalNodeJS{},
		mu:               new(sync.Mutex),
		workerAddress:    cfg.WorkerAddress,
		cfg:              cfg,
	}

	configuredRunnerType = runnerTypeLocal
	runnerManager = lm
	return nil
}

func (l *localRunnerManager) Start(_ context.Context, sessionID sdktypes.SessionID, buildArtifacts []byte, vars map[string]string) (string, *RunnerClient, error) {
	log := l.logger.With(zap.String("session_id", sessionID.String()))
	r := &LocalNodeJS{
		log:           log,
		logRunnerCode: l.cfg.LogCodeRunnerCode,
		sessionID:     sessionID,
	}

	if l.workerAddress == "" {
		l.workerAddress = l.cfg.WorkerAddressProvider()
		if l.workerAddress == "" {
			log.Error("worker address could not be set")
			return "", nil, errors.New("worker address wasnt provided and could not be inferred")
		}

		log.Info("worker address inferred", zap.String("workerAddress", l.workerAddress))
	}

	if err := r.Start(l.nodeExe, buildArtifacts, vars, l.workerAddress); err != nil {
		return "", nil, err
	}

	runnerAddr := fmt.Sprintf("127.0.0.1:%d", r.port)
	log.Debug("dialing runner", zap.String("addr", runnerAddr))
	client, err := dialRunner(runnerAddr)
	if err != nil {
		if err := r.Close(); err != nil {
			log.Warn("close runner", zap.Error(err))
		}
		return "", nil, err
	}

	l.mu.Lock()
	l.runnerIDToRunner[r.id] = r
	l.mu.Unlock()
	return r.id, client, nil
}

func (l *localRunnerManager) RunnerHealth(ctx context.Context, runnerID string) error {
	l.mu.Lock()
	runner, ok := l.runnerIDToRunner[runnerID]
	l.mu.Unlock()

	if !ok {
		return errors.New("runner not found")
	}

	return runner.Health()
}

func (l *localRunnerManager) Stop(ctx context.Context, runnerID string) error {
	l.mu.Lock()
	runner, ok := l.runnerIDToRunner[runnerID]
	l.mu.Unlock()

	if !ok {
		return errors.New("not found")
	}

	err := runner.Close()

	l.mu.Lock()
	delete(l.runnerIDToRunner, runnerID)
	l.mu.Unlock()

	return err
}

func (*localRunnerManager) Health(ctx context.Context) error { return nil }
