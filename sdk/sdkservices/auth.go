package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Auth interface {
	WhoAmI(ctx context.Context) (sdktypes.User, error)
	CreateToken(ctx context.Context) (string, error)
}
