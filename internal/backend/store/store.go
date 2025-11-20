package store

import (
	"context"
	"errors"
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
	MaxValueSizeBytes   int `koanf:"max_value_size_bytes"`
	MaxValuesPerProject int `koanf:"max_values_per_project"`
}

var Configs = configset.Set[Config]{
	Default: &Config{
		MaxValueSizeBytes:   64 * 1024, // 64 KiB
		MaxValuesPerProject: 64,
	},
}

type store struct {
	cfg *Config
	db  db.DB
	l   *zap.Logger
}

func New(db db.DB, l *zap.Logger, cfg *Config) sdkservices.Store {
	return &store{db: db, l: l, cfg: cfg}
}

func (s *store) Mutate(ctx context.Context, pid sdktypes.ProjectID, key, op string, operands ...sdktypes.Value) (sdktypes.Value, error) {
	ret := sdktypes.Nothing

	if err := s.db.Transaction(ctx, func(tx db.DB) error {
		r, ok := ops[op]
		if !ok {
			return sdkerrors.NewInvalidArgumentError("unknown operation")
		}

		var (
			curr, next sdktypes.Value
			err        error
		)

		if r.read {
			if err := authz.CheckContext(
				ctx,
				pid,
				authz.OpStoreReadGet,
				authz.WithData("keys", []string{key}),
				authz.WithConvertForbiddenToNotFound,
			); err != nil {
				return err
			}

			curr, err = tx.GetStoreValue(ctx, pid, key)
			if err != nil && !errors.Is(err, sdkerrors.ErrNotFound) {
				return err
			}
		}

		if r.fn != nil {
			if next, ret, err = r.fn(curr, operands); err != nil {
				return err
			}
		}

		if r.read && curr.Equal(next) {
			return nil
		}

		if r.write {
			if next.ProtoSize() > s.cfg.MaxValueSizeBytes {
				return sdkerrors.NewInvalidArgumentError("value size (%d bytes) exceeds maximum allowed (%d bytes)", next.ProtoSize(), s.cfg.MaxValueSizeBytes)
			}

			if !curr.IsValid() {
				count, err := tx.CountStoreValues(ctx, pid)
				if err != nil {
					return err
				}

				if int(count)+1 > s.cfg.MaxValuesPerProject {
					return sdkerrors.NewInvalidArgumentError("number of stored values (%d) exceeds maximum allowed (%d)", count+1, s.cfg.MaxValuesPerProject)
				}
			}

			if err := authz.CheckContext(
				ctx,
				pid,
				authz.OpStoreWriteSet,
				authz.WithData("key", key),
				authz.WithData("size", next.ProtoSize()),
			); err != nil {
				// TODO: Error here would be forbidden if size is too large. Figure out how to communicate that to the user.
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
		authz.OpStoreReadGet,
		authz.WithData("keys", keys),
		authz.WithConvertForbiddenToNotFound,
	); err != nil {
		return nil, err
	}

	return s.db.ListStoreValues(ctx, pid, keys, true)
}

func (s *store) List(ctx context.Context, pid sdktypes.ProjectID) ([]string, error) {
	if err := authz.CheckContext(ctx, pid, authz.OpStoreReadList); err != nil {
		return nil, err
	}

	vs, err := s.db.ListStoreValues(ctx, pid, nil, false)
	if err != nil {
		return nil, err
	}

	return slices.Sorted(maps.Keys(vs)), nil
}

func (s *store) Publish(ctx context.Context, pid sdktypes.ProjectID, key string) error {
	if err := authz.CheckContext(
		ctx,
		pid,
		authz.OpStoreWritePublish,
		authz.WithData("key", key),
	); err != nil {
		return err
	}

	return s.db.PublishStoreValue(ctx, pid, key)
}

func (s *store) Unpublish(ctx context.Context, pid sdktypes.ProjectID, key string) error {
	if err := authz.CheckContext(
		ctx,
		pid,
		authz.OpStoreWriteUnpublish,
		authz.WithData("key", key),
	); err != nil {
		return err
	}

	return s.db.UnpublishStoreValue(ctx, pid, key)
}
