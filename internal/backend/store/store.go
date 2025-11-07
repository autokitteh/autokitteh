package store

import (
	"context"
	"errors"
	"maps"
	"slices"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type store struct {
	db db.DB
	l  *zap.Logger
}

func New(db db.DB, l *zap.Logger) sdkservices.Store {
	return &store{db: db, l: l}
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
		authz.WithConvertForbiddenToNotFound,
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
