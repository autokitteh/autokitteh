package envs

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/db"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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
	if sdktypes.GetEnvProjectID(env) == nil {
		return nil, fmt.Errorf("%w: missing project ID", sdkerrors.ErrInvalidArgument)
	}

	env, err := env.Update(func(pb *sdktypes.EnvPB) {
		pb.EnvId = sdktypes.NewEnvID().String()
	})
	if err != nil {
		return nil, err
	}

	if sdktypes.GetEnvProjectID(env) == nil {
		return nil, fmt.Errorf("%w: owner must be specified", sdkerrors.ErrInvalidArgument)
	}

	if env, err = sdktypes.ToStrictEnv(env); err != nil {
		return nil, err
	}

	if err := e.db.CreateEnv(ctx, env); err != nil {
		return nil, err
	}

	return sdktypes.GetEnvID(env), nil
}

func (e *envs) GetByID(ctx context.Context, eid sdktypes.EnvID) (sdktypes.Env, error) {
	return sdkerrors.IgnoreNotFoundErr(e.db.GetEnvByID(ctx, eid))
}

func (e *envs) GetByName(ctx context.Context, pid sdktypes.ProjectID, en sdktypes.Name) (sdktypes.Env, error) {
	if pid == nil {
		return nil, fmt.Errorf("%w: missing project ID", sdkerrors.ErrInvalidArgument)
	}

	return sdkerrors.IgnoreNotFoundErr(e.db.GetEnvByName(ctx, pid, en))
}

func (e *envs) List(ctx context.Context, pid sdktypes.ProjectID) ([]sdktypes.Env, error) {
	return e.db.ListProjectEnvs(ctx, pid)
}

func (e *envs) SetVar(ctx context.Context, ev sdktypes.EnvVar) error {
	return e.db.SetEnvVar(ctx, ev)
}

func (e *envs) GetVars(ctx context.Context, vns []sdktypes.Symbol, eid sdktypes.EnvID) ([]sdktypes.EnvVar, error) {
	vs, err := e.db.GetEnvVars(ctx, eid)

	if len(vns) != 0 {
		has := kittehs.ContainedIn(kittehs.TransformToStrings(vns)...)

		vs = kittehs.Filter(
			vs,
			func(v sdktypes.EnvVar) bool { return has(sdktypes.GetEnvVarName(v).String()) },
		)
	}

	return vs, err
}

func (e *envs) RevealVar(ctx context.Context, eid sdktypes.EnvID, vn sdktypes.Symbol) (string, error) {
	return e.db.RevealEnvVar(ctx, eid, vn)
}

// TODO
func (e *envs) Remove(context.Context, sdktypes.EnvID) error                     { return nil }
func (e *envs) Update(context.Context, sdktypes.Env) error                       { return nil }
func (e *envs) RemoveVar(context.Context, sdktypes.EnvID, sdktypes.Symbol) error { return nil }
