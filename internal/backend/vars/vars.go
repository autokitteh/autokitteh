package vars

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/secrets"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type vars struct {
	db      db.DB
	secrets secrets.Secrets
	z       *zap.Logger
}

func New(z *zap.Logger, db db.DB, secrets secrets.Secrets) sdkservices.Vars {
	return &vars{db: db, z: z, secrets: secrets}
}

func varSecretKey(secret sdktypes.Var) string {
	return fmt.Sprintf("%s/%s", secret.ScopeID().AsID().UUIDValue().String(), secret.Name())
}

func (v *vars) Set(ctx context.Context, vs ...sdktypes.Var) error {
	for i, va := range vs {
		if va.IsSecret() {
			key := varSecretKey(va)
			if err := v.secrets.Set(ctx, key, map[string]string{"value": va.Value()}); err != nil {
				//TODO: handle delete secrets that were already created
				return err
			}

			vs[i] = sdktypes.NewVar(va.Name(), key, true).WithScopeID(va.ScopeID())
		}
	}
	return v.db.SetVars(ctx, vs)
}

func (v *vars) Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error {
	vars, err := v.db.GetVars(ctx, sid, names)
	if err != nil {
		return err
	}

	if err := v.db.DeleteVars(ctx, sid, names); err != nil {
		return err
	}

	for _, va := range vars {
		if va.IsSecret() {
			key := varSecretKey(va)
			err = v.secrets.Delete(ctx, key)
			if err != nil {
				v.z.Error("failed delete secret", zap.String("key", key), zap.Error(err))
			}
		}
	}

	return err
}

func (v *vars) Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	return v.db.GetVars(ctx, sid, names)
}

func (v *vars) Reveal(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	vars, err := v.db.GetVars(ctx, sid, names)
	if err != nil {
		return nil, err
	}

	var resultVars sdktypes.Vars
	for _, va := range vars {
		key := varSecretKey(va)
		value, err := v.secrets.Get(ctx, key)
		if err != nil {
			return nil, err
		}

		data, err := json.Marshal(value["value"])
		if err != nil {
			return nil, err
		}

		resultVars = append(resultVars, sdktypes.NewVar(va.Name(), string(data), true))
	}

	return resultVars, nil
}

func (v *vars) FindConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	return v.db.FindConnectionIDsByVar(ctx, iid, name, value)
}
