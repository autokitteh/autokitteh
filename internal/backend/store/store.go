package store

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Config struct {
	MaxValueSize           int `koanf:"max_value_size"`             // 0 to disable size limit.
	MaxStoreKeysPerProject int `koanf:"max_store_keys_per_project"` // 0 to disable limit.
}

var Configs = configset.Set[Config]{
	Default: &Config{
		MaxValueSize:           64 * 1024, // 64kb
		MaxStoreKeysPerProject: 64,
	},
}

type store struct {
	db  db.DB
	cfg *Config
	l   *zap.Logger
}

func New(cfg *Config, db db.DB, l *zap.Logger) sdkservices.Store {
	return &store{db: db, l: l, cfg: cfg}
}

func (s *store) enforceLimits(
	ctx context.Context,
	tx db.TX,
	op op,
	curr, next sdktypes.Value,
	pid sdktypes.ProjectID,
	key string,
) error {
	if s.cfg.MaxValueSize != 0 && op.write && next.ProtoSize() > s.cfg.MaxValueSize {
		return sdkerrors.NewInvalidArgumentError("value size %d exceeds maximum allowed size %d for a single value", next.ProtoSize(), s.cfg.MaxValueSize)
	}

	if maxCount := int64(s.cfg.MaxStoreKeysPerProject); maxCount > 0 {
		isNewKey := false
		if op.read {
			isNewKey = curr.IsNothing()
		} else {
			has, err := tx.HasStoreKey(ctx, pid, key)
			if err != nil {
				return fmt.Errorf("has: %w", err)
			}

			isNewKey = !has
		}

		if isNewKey {
			count, err := tx.CountStoreKeys(ctx, pid)
			if err != nil {
				return fmt.Errorf("count: %w", err)
			}

			if count >= maxCount {
				return sdkerrors.NewInvalidArgumentError("maximum number of store keys (%d) reached for project", maxCount)
			}
		}
	}

	return nil
}

func (s *store) Mutate(ctx context.Context, pid sdktypes.ProjectID, key, op string, operands ...sdktypes.Value) (sdktypes.Value, error) {
	if err := authz.CheckContext(
		ctx,
		pid,
		"write:do",
		authz.WithData("op", op),
	); err != nil {
		return sdktypes.InvalidValue, err
	}

	ret := sdktypes.Nothing

	r, ok := ops[op]
	if !ok {
		return sdktypes.InvalidValue, sdkerrors.NewInvalidArgumentError("unknown operation")
	}

	if err := s.db.Transaction(ctx, func(tx db.TX) error {
		if r.write && s.cfg.MaxValueSize != 0 {
			// The lock is only from R-M-W and concerning only values count currently.
			if err := tx.LockProject(ctx, pid); err != nil {
				return fmt.Errorf("lock: %w", err)
			}
		}

		var curr, next sdktypes.Value

		if r.read {
			var err error
			curr, err = tx.GetStoreValue(ctx, pid, key)
			if err != nil && !errors.Is(err, sdkerrors.ErrNotFound) {
				return err
			}
		}

		if r.fn != nil {
			var err error
			if next, ret, err = r.fn(curr, operands); err != nil {
				return err
			}
		}

		if r.read && curr.Equal(next) {
			return nil
		}

		if r.write {
			if err := s.enforceLimits(ctx, tx, r, curr, next, pid, key); err != nil {
				return err
			}

			if err := tx.SetStoreValue(ctx, pid, key, next); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return sdktypes.InvalidValue, err
	}

	return ret, nil
}

func (s *store) Get(ctx context.Context, pid sdktypes.ProjectID, keys []string) (map[string]sdktypes.Value, error) {
	if err := authz.CheckContext(
		ctx,
		pid,
		"read:get",
		authz.WithData("keys", keys),
	); err != nil {
		return nil, err
	}

	return s.db.ListStoreValues(ctx, pid, keys, true)
}

func (s *store) List(ctx context.Context, pid sdktypes.ProjectID) ([]string, error) {
	if err := authz.CheckContext(ctx, pid, "read:list"); err != nil {
		return nil, err
	}

	vs, err := s.db.ListStoreValues(ctx, pid, nil, false)
	if err != nil {
		return nil, err
	}

	return slices.Sorted(maps.Keys(vs)), nil
}
