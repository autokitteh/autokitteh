package envsauth

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/auth/authcontext"
)

type envs struct {
	envs sdkservices.Envs
}

func Wrap(in sdkservices.Envs) sdkservices.Envs { return &envs{envs: in} }

func (e *envs) Create(ctx context.Context, env sdktypes.Env) (sdktypes.EnvID, error) {
	if userID := authcontext.GetAuthnUserID(ctx); !env.ParentID().IsValid() && userID.IsValid() {
		env = env.WithParentID(userID)
	}

	return e.envs.Create(ctx, env)
}

func (e *envs) GetByID(ctx context.Context, eid sdktypes.EnvID) (sdktypes.Env, error) {
	return e.envs.GetByID(ctx, eid)
}

func (e *envs) GetByName(ctx context.Context, epid sdktypes.EnvParentID, en sdktypes.Name) (sdktypes.Env, error) {
	if userID := authcontext.GetAuthnUserID(ctx); epid == nil && userID != nil {
		epid = sdktypes.NewEnvParentID(userID)
	}

	return e.envs.GetByName(ctx, epid, en)
}

func (e *envs) List(ctx context.Context, epid sdktypes.EnvParentID) ([]sdktypes.Env, error) {
	if userID := authcontext.GetAuthnUserID(ctx); epid == nil && userID != nil {
		epid = sdktypes.NewEnvParentID(userID)
	}

	return e.envs.List(ctx, epid)
}

func (e *envs) SetVar(ctx context.Context, ev sdktypes.EnvVar) error {
	return e.envs.SetVar(ctx, ev)
}

func (e *envs) GetVars(ctx context.Context, vns []sdktypes.Symbol, eid sdktypes.EnvID) ([]sdktypes.EnvVar, error) {
	return e.envs.GetVars(ctx, vns, eid)
}

func (e *envs) RevealVar(ctx context.Context, eid sdktypes.EnvID, vn sdktypes.Symbol) (string, error) {
	return e.envs.RevealVar(ctx, eid, vn)
}

func (e *envs) Remove(ctx context.Context, eid sdktypes.EnvID) error {
	return e.envs.Remove(ctx, eid)
}

func (e *envs) Update(ctx context.Context, env sdktypes.Env) error {
	return e.envs.Update(ctx, env)
}

func (e *envs) RemoveVar(ctx context.Context, eid sdktypes.EnvID, n sdktypes.Symbol) error {
	return e.envs.RemoveVar(ctx, eid, n)
}
