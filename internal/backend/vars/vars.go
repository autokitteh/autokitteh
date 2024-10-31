package vars

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/secrets"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	MaxValueSize       int `koanf:"max_value_size"`
	MaxNumVarsPerScope int `koanf:"max_num_vars_per_scope"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		MaxValueSize:       8092,
		MaxNumVarsPerScope: 64,
	},
}

type Vars struct {
	cfg     Config
	db      db.DB
	secrets secrets.Secrets
	conns   sdkservices.Connections
	z       *zap.Logger
}

func New(z *zap.Logger, cfg *Config, db db.DB, secrets secrets.Secrets) *Vars {
	return &Vars{db: db, z: z, secrets: secrets, cfg: *cfg}
}

func varSecretKey(secret sdktypes.Var) string {
	return fmt.Sprintf("%s/%s", secret.ScopeID().AsID().UUIDValue().String(), secret.Name())
}

func (v *Vars) SetConnections(conns sdkservices.Connections) { v.conns = conns }

// Set sets the given variables in the database. If any of the variables are
// secrets, it stores them in the secret store. Note that this function modifies
// the values of secret variables to be the secret key in the secret store.
// Do not change this behavior - it's useful, even though it's unexpected.
func (v *Vars) Set(ctx context.Context, vs ...sdktypes.Var) error {
	scids := make(map[sdktypes.VarScopeID]bool, len(vs))

	for i, va := range vs {
		if va.IsSecret() {
			key := varSecretKey(va)
			if err := v.secrets.Set(ctx, key, va.Value()); err != nil {
				// TODO: ENG-817 - handle dangling secrets in secret store
				return err
			}

			vs[i] = va.SetValue(key)
		}

		if v.cfg.MaxValueSize != 0 && len(va.Value()) > v.cfg.MaxValueSize {
			return fmt.Errorf("%w: value size %d exceeds max value size %d", sdkerrors.ErrLimitExceeded, len(va.Value()), v.cfg.MaxValueSize)
		}

		scids[va.ScopeID()] = true
	}

	err := v.db.Transaction(ctx, func(tx db.DB) error {
		if maxN := v.cfg.MaxNumVarsPerScope; maxN != 0 && len(vs) > maxN {
			for scid := range scids {
				n, err := tx.CountVars(ctx, scid)
				if err != nil {
					return err
				}

				if n > maxN {
					return fmt.Errorf("%w: number of variables for scope %v %d exceeds max number of variables %d", sdkerrors.ErrLimitExceeded, scid, len(vs), maxN)
				}
			}
		}

		return tx.SetVars(ctx, vs)
	})
	if err != nil {
		return err
	}

	sids := kittehs.ListToMap(vs, func(v sdktypes.Var) (sdktypes.VarScopeID, bool) { return v.ScopeID(), v.ScopeID().IsConnectionID() })
	var errs []error
	for sid, isConn := range sids {
		if isConn {
			if _, err := v.conns.RefreshStatus(ctx, sid.ToConnectionID()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (v *Vars) Delete(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) error {
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

	if sid.IsConnectionID() {
		if _, err := v.conns.RefreshStatus(ctx, sid.ToConnectionID()); err != nil {
			return err
		}
	}

	return nil
}

func (v *Vars) Get(ctx context.Context, sid sdktypes.VarScopeID, names ...sdktypes.Symbol) (sdktypes.Vars, error) {
	vars, err := v.db.GetVars(ctx, sid, names)
	if err != nil {
		return nil, err
	}

	return kittehs.TransformError(vars, func(va sdktypes.Var) (sdktypes.Var, error) {
		if !va.IsSecret() {
			return va, nil
		}

		key := varSecretKey(va)
		value, err := v.secrets.Get(ctx, key)
		if err != nil {
			return sdktypes.InvalidVar, err
		}

		return va.SetValue(value), nil
	})
}

func (v *Vars) FindConnectionIDs(ctx context.Context, iid sdktypes.IntegrationID, name sdktypes.Symbol, value string) ([]sdktypes.ConnectionID, error) {
	return v.db.FindConnectionIDsByVar(ctx, iid, name, value)
}
