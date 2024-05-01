package vars

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type vars struct{ db db.DB }

func New(z *zap.Logger, db db.DB) sdkservices.Vars {
	return &vars{db: db}
}

func (v *vars) Set(ctx context.Context, vs ...sdktypes.Var) error {
	return v.db.SetVars(ctx, vs)
}

func (v *vars) Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error {
	return v.db.DeleteVars(ctx, sid, names)
}

func (v *vars) Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	return v.db.GetVars(ctx, false, sid, names)
}

func (v *vars) Reveal(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	return v.db.GetVars(ctx, true, sid, names)
}

func (v *vars) FindConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	return v.db.FindConnectionIDsByVar(ctx, iid, name, value)
}
