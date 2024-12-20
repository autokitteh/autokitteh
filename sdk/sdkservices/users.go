package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Users interface {
	Create(ctx context.Context, user sdktypes.User) (sdktypes.UserID, error)

	Update(ctx context.Context, user sdktypes.User) error

	// at least one of the arguments must be non-zero.
	Get(ctx context.Context, id sdktypes.UserID, name sdktypes.Symbol, email string) (sdktypes.User, error)
}
