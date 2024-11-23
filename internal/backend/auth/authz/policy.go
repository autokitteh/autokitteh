package authz

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
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
	return func(ctx context.Context, id sdktypes.ID, action string, opts ...func(*checkCfg)) error {
		var cfg checkCfg
		for _, opt := range opts {
			opt(&cfg)
		}

		input, err := buildInput(ctx, db, id, action, cfg)
		if err != nil {
			return err
		}

		result, err := decide(ctx, policyRootPath+"allow", input)
		if err != nil {
			return fmt.Errorf("authz opa decision: %w", err)
		}

		decision, ok := result.(bool)
		if !ok {
			return fmt.Errorf("authz opa decision: not a boolean")
		}

		if !decision {
			l.Warn("authz opa decision: denied", zap.Any("input", input), zap.Any("result", result))
			return sdkerrors.ErrUnauthorized
		}

		return nil
	}
}

func buildInput(ctx context.Context, db db.DB, id sdktypes.ID, action string, cfg checkCfg) (map[string]any, error) {
	var oid sdktypes.OwnerID

	if id.IsValid() {
		var err error
		if oid, err = db.GetOwner(ctx, id); err != nil {
			return nil, fmt.Errorf("get owner: %w", err)
		}
	}

	if pid := cfg.belongsToProjectOfID; pid != nil && pid.IsValid() {
		oid, err := db.GetOwner(ctx, pid)
		if err != nil {
			return nil, fmt.Errorf("get project owner: %w", err)
		}

		cfg.data["project_owner_id"] = oid.String()
	}

	uid := authcontext.GetAuthnUserID(ctx)

	actType, act, ok := strings.Cut(action, ":")
	if !ok {
		actType, act = "", action
	}

	return map[string]any{
		"kind":        id.Kind(),
		"action_type": actType,
		"action":      act,
		"data":        cfg.data,
		"user_id":     uid.String(), // requester user id.
		"owner_id":    oid.String(), // owner of the resource. empty if no specific resource.
		"resource_id": id.String(),  // empty if no specific resource.
	}, nil
}
