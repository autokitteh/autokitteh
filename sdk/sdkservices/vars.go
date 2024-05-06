package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Vars interface {
	Set(ctx context.Context, v ...sdktypes.Var) error
	Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error
	Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error)
	Reveal(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error)
	FindConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error)
}
