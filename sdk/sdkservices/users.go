package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Users interface {
	Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error)
	FindByProvider(ctx context.Context, p sdktypes.UserAuthProvider) (sdktypes.UserID, error)
	FindByProviderOrCreate(ctx context.Context, p sdktypes.UserAuthProvider) (_ sdktypes.UserID, created bool, _ error)
	GetByID(ctx context.Context, id sdktypes.UserID) (sdktypes.User, error)
}
