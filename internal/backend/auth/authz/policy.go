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
	"go.autokitteh.dev/autokitteh/internal/kittehs"
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

		result, err := decide(ctx, "authz/allow", input)
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
	uid := authcontext.GetAuthnUserID(ctx)

	uoids, err := db.GetOrgsForUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get orgs for user: %w", err)
	}

	actType, act, ok := strings.Cut(action, ":")
	if !ok {
		actType, act = "", action
	}

	data := map[string]any{
		"kind":         id.Kind(),                         // resource kind. available even if id is invalid, it still contains the kind.
		"user_id":      uid.String(),                      // requester user id.
		"user_org_ids": kittehs.TransformToStrings(uoids), // orgs the user is part of.
		"action_type":  actType,                           // [type:]xxx of action, or "" if not specified.
		"action":       act,                               // [xxx:]action of action.
		"data":         cfg.data,                          // aux data supplied by the caller.
	}

	if id.IsValid() {
		// Get the resource owner.

		owid, err := db.GetOwner(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get owner: %w", err)
		}

		if !owid.IsValid() {
			return nil, fmt.Errorf("could not figure out owner of the resource")
		}

		data["resource_owner_id"] = owid.String()
	}

	for name, id := range cfg.associations {
		// In case the resource is associated with a something, we need to get that thing's owner.
		// This is relevant, for example, with new builds or sessions, where they are explicitly owned
		// but can be associated with a project. The policy needs to decide if to allow it.

		owid, err := db.GetOwner(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get project owner: %w", err)
		}

		data[fmt.Sprintf("associated_%s_owner_id", name)] = owid.String()
	}

	return data, nil
}
