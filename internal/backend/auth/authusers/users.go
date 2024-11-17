package authusers

import (
	"context"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Users interface {
	Create(ctx context.Context, u sdktypes.User) (sdktypes.UserID, error)
	GetByID(ctx context.Context, id sdktypes.UserID) (sdktypes.User, error)
	GetByEmail(ctx context.Context, email string) (sdktypes.User, error)
}

type users struct {
	db db.DB
	l  *zap.Logger
}

func New(db db.DB, l *zap.Logger) Users {
	if l == nil {
		l = zap.NewNop()
	}

	return &users{db: db, l: l}
}

func (us *users) Create(ctx context.Context, u sdktypes.User) (sdktypes.UserID, error) {
	return us.db.CreateUser(ctx, u)
}

func (us *users) GetByID(ctx context.Context, id sdktypes.UserID) (sdktypes.User, error) {
	return us.db.GetUserByID(ctx, id)
}

func (us *users) GetByEmail(ctx context.Context, email string) (sdktypes.User, error) {
	return us.db.GetUserByEmail(ctx, email)
}
