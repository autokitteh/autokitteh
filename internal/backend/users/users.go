package users

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/auth/authz"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type users struct {
	db db.DB
	l  *zap.Logger
}

func New(db db.DB, l *zap.Logger) sdkservices.Users {
	return &users{db: db, l: l}
}

func (u *users) Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error) {
	if user.ID().IsValid() {
		return sdktypes.InvalidUserID, sdkerrors.NewInvalidArgumentError("user ID must be empty")
	}

	if err := authz.CheckContext(ctx, sdktypes.InvalidUserID, "create:create", authz.WithData("user", user)); err != nil {
		return sdktypes.InvalidUserID, err
	}

	return u.db.CreateUser(ctx, user)
}

func (u *users) Get(ctx context.Context, id sdktypes.UserID, email string) (sdktypes.User, error) {
	if email == "" && !id.IsValid() {
		id = authcontext.GetAuthnUserID(ctx)
	}

	if err := authz.CheckContext(ctx, id, "read:get", authz.WithData("id", id.String()), authz.WithData("email", email)); err != nil {
		return sdktypes.InvalidUser, err
	}

	return u.db.GetUser(ctx, id, email)
}

func (u *users) Update(ctx context.Context, user sdktypes.User) error {
	if !user.ID().IsValid() {
		return sdkerrors.NewInvalidArgumentError("missing user ID")
	}

	if err := authz.CheckContext(ctx, user.ID(), "update:update", authz.WithData("user", user)); err != nil {
		return err
	}

	return u.db.UpdateUser(ctx, user)
}
