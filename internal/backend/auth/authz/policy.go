package authz

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/policy"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

// PolicyCheckFunc is a function that checks access to a resource using a policy.
// Actions can be either of:
//   - "some_action_name"             -> {"action": "some_action_name", "action_type": ""}
//   - "action_type:some_action_name" -> {"action": "some_action_name", "action_type": "action_type"}
func NewPolicyCheckFunc(l *zap.Logger, db db.DB, decide policy.DecideFunc) CheckFunc {
	return func(ctx context.Context, id sdktypes.ID, action string, opts ...CheckOpt) error {
		cfg := configure(opts)

		input, err := buildInput(ctx, db, id, action, cfg)
		if err != nil {
			return err
		}

		result, err := decide(ctx, "authz/allow", input)
		if err != nil {
			return fmt.Errorf("authz opa decision: %w", err)
		}

		decision, ok := result.(bool)
		if !ok {
			return errors.New("authz opa decision: not a boolean")
		}

		l := l.With(zap.Any("input", input), zap.Any("result", result))
		l.Debug("authz opa decision")

		if !decision {
			l.WithOptions(zap.AddStacktrace(zap.WarnLevel)).Warn("authz opa decision: denied")

			if cfg.convertForbiddenToNotFound {
				return sdkerrors.ErrNotFound
			}

			return sdkerrors.ErrUnauthorized
		}

		return nil
	}
}
