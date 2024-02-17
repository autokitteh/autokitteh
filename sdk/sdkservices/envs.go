package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Envs interface {
	List(ctx context.Context, projectID sdktypes.ProjectID) ([]sdktypes.Env, error)
	Create(ctx context.Context, env sdktypes.Env) (sdktypes.EnvID, error)
	GetByID(ctx context.Context, envID sdktypes.EnvID) (sdktypes.Env, error)
	GetByName(ctx context.Context, projectID sdktypes.ProjectID, name sdktypes.Name) (sdktypes.Env, error)
	Remove(ctx context.Context, envID sdktypes.EnvID) error
	Update(ctx context.Context, env sdktypes.Env) error

	SetVar(ctx context.Context, envVar sdktypes.EnvVar) error
	RemoveVar(ctx context.Context, envID sdktypes.EnvID, name sdktypes.Symbol) error
	GetVars(ctx context.Context, names []sdktypes.Symbol, envID sdktypes.EnvID) ([]sdktypes.EnvVar, error)
	RevealVar(ctx context.Context, envID sdktypes.EnvID, varName sdktypes.Symbol) (string, error)
}
