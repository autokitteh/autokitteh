package authz

import (
	"context"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ctxKey string

var checkFuncKey = ctxKey("authz-check-func")

// Called if no check function is set in the context.
var defaultCheck = func(context.Context, sdktypes.ID, string, ...func(*checkCfg)) error {
	return fmt.Errorf("missing authz in context: %w", sdkerrors.ErrNotImplemented)
}

func DisableCheckForTesting() {
	defaultCheck = func(context.Context, sdktypes.ID, string, ...func(*checkCfg)) error { return nil }
}

func ContextWithCheckFunc(ctx context.Context, check CheckFunc) context.Context {
	return context.WithValue(ctx, checkFuncKey, check)
}

func getContextCheckFunc(ctx context.Context) CheckFunc {
	if check, ok := ctx.Value(checkFuncKey).(CheckFunc); ok {
		return check
	}
	return nil
}

func CheckContext(ctx context.Context, id sdktypes.ID, action string, opts ...func(*checkCfg)) error {
	if authcontext.IsAuthnSystemUser(ctx) {
		// When system user is specified there is usually no authz check in context, so we should
		// just make it all approved here.
		return nil
	}

	check := getContextCheckFunc(ctx)
	if check == nil {
		check = defaultCheck
	}

	return check(ctx, id, action, opts...)
}
