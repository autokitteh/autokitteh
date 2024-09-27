package sessionworkflows

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (w *sessionWorkflow) refreshGoogleOAuth(ctx context.Context, args []sdktypes.Value, kwargs map[string]sdktypes.Value) (sdktypes.Value, error) {
	w.l.Warn("Google OAuth refresh", zap.Any("unwrap", args), zap.Any("unwrap", kwargs))

	return sdktypes.InvalidValue, errors.New("not implemented")
}
