package envs

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type envs struct {
	z  *zap.Logger
	db db.DB
}

func New(z *zap.Logger, db db.DB) sdkservices.Envs {
	return &envs{db: db, z: z}
}

func (e *envs) Create(ctx context.Context, env sdktypes.Env) (sdktypes.EnvID, error) {
	if !env.ProjectID().IsValid() {
		return sdktypes.InvalidEnvID, sdkerrors.NewInvalidArgumentError("missing project ID")
	}

	env = env.WithNewID()

	if err := e.db.CreateEnv(ctx, env); err != nil {
		return sdktypes.InvalidEnvID, err
	}

	return env.ID(), nil
}

func (e *envs) GetByID(ctx context.Context, eid sdktypes.EnvID) (sdktypes.Env, error) {
	return sdkerrors.IgnoreNotFoundErr(e.db.GetEnvByID(ctx, eid))
}

func (e *envs) GetByName(ctx context.Context, pid sdktypes.ProjectID, en sdktypes.Symbol) (sdktypes.Env, error) {
	if !pid.IsValid() {
		return sdktypes.InvalidEnv, sdkerrors.NewInvalidArgumentError("missing project ID")
	}

	return sdkerrors.IgnoreNotFoundErr(e.db.GetEnvByName(ctx, pid, en))
}

func (e *envs) List(ctx context.Context, pid sdktypes.ProjectID) ([]sdktypes.Env, error) {
	return e.db.ListProjectEnvs(ctx, pid)
}

// TODO
func (e *envs) Remove(context.Context, sdktypes.EnvID) error { return nil }
func (e *envs) Update(context.Context, sdktypes.Env) error   { return nil }
