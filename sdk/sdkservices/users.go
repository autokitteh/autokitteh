package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Users interface {
	Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error)
	GetByID(ctx context.Context, userID sdktypes.UserID) (sdktypes.User, error)
	GetByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.User, error)
}
