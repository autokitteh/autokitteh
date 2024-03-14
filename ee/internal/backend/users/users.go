package users

import (
	"context"
	"errors"
	"fmt"

	"connectrpc.com/connect"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"

	"go.autokitteh.dev/autokitteh/ee/internal/backend/db"
)

type Users struct {
	fx.In

	Z  *zap.Logger
	DB db.DB
}

func New(u Users) sdkservices.Users { return &u }

func (o *Users) Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error) {
	tx, err := o.DB.Begin(ctx)
	if err != nil {
		return sdktypes.InvalidUserID, fmt.Errorf("db.begin: %w", err)
	}

	defer db.LoggedRollback(o.Z, tx)

	if _, err := o.DB.GetOrgByName(ctx, user.Name()); err == nil {
		return sdktypes.InvalidUserID, fmt.Errorf("%w: an org with the same name already exists", sdkerrors.ErrAlreadyExists)
	} else if !errors.Is(err, sdkerrors.ErrNotFound) {
		return sdktypes.InvalidUserID, fmt.Errorf("orgs.get: %w", err)
	}

	user = user.WithNewID()

	if err := user.Strict(); err != nil {
		return sdktypes.InvalidUserID, err
	}

	if err := tx.CreateUser(ctx, user); err != nil {
		if errors.Is(err, sdkerrors.ErrAlreadyExists) {
			return sdktypes.InvalidUserID, connect.NewError(connect.CodeAlreadyExists, err)
		}

		return sdktypes.InvalidUserID, connect.NewError(connect.CodeUnknown, err)
	}

	if err := tx.Commit(); err != nil {
		return sdktypes.InvalidUserID, fmt.Errorf("db.commit: %w", err)
	}

	return user.ID(), nil
}

func (o *Users) GetByID(ctx context.Context, uid sdktypes.UserID) (sdktypes.User, error) {
	return sdkerrors.IgnoreNotFoundErr(o.DB.GetUserByID(ctx, uid))
}

func (o *Users) GetByName(ctx context.Context, h sdktypes.Symbol) (sdktypes.User, error) {
	return sdkerrors.IgnoreNotFoundErr(o.DB.GetUserByName(ctx, h))
}
