//go:build !enterprise
// +build !enterprise

package workflowexecutor

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	SessionWorkflow temporalclient.WorkflowConfig `koanf:"session_workflow"`
}

var (
	Configs = configset.Set[Config]{
		Default: &Config{},
	}
)

var _ WorkflowExecutor = (*executor)(nil)

type executor struct {
	svcs Svcs
	l    *zap.Logger
	cfg  *Config
}

func New(svcs Svcs, l *zap.Logger, cfg *Config) (*executor, error) {
	return &executor{svcs: svcs, l: l, cfg: cfg}, nil
}

func (e *executor) WorkflowQueue() string {
	return "sessions"
}

func (e *executor) Execute(ctx context.Context, sessionID sdktypes.SessionID, data any, memo map[string]string) error {
	return e.execute(ctx, sessionID, workflowID(sessionID), data, memo)
}

func (e *executor) NotifyDone(ctx context.Context, id string) error {
	return nil
}

func (e *executor) Start(ctx context.Context) error {

	return nil
}

func (e *executor) Stop(ctx context.Context) error {

	return nil
}
