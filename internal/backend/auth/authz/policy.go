package authz

import (
	"context"
	_ "embed"
	"fmt"
	"maps"
	"slices"
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

			if cfg.convertForbiddenToNotFound {
				return sdkerrors.ErrNotFound
			}

			return sdkerrors.ErrUnauthorized
		}

		return nil
	}
}

func buildInput(ctx context.Context, db db.DB, id sdktypes.ID, action string, cfg checkCfg) (map[string]any, error) {
	uid := authcontext.GetAuthnUserID(ctx)

	uorgs, err := db.GetOrgsForUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("get orgs for user: %w", err)
	}

	uoids := kittehs.Transform(uorgs, func(org sdktypes.Org) string { return org.ID().String() })

	// TODO: Use https://docs.styra.com/opa/rego-by-example/builtins/regex/globs_match to match permissions.
	actType, act, ok := strings.Cut(action, ":")
	if !ok {
		actType, act = "", action
	}

	oidsSet := make(map[string]bool)
	pidsSet := make(map[string]bool)

	var (
		oid sdktypes.OrgID
		pid sdktypes.ProjectID
	)

	if id.IsValid() {
		// Get the resource org.
		if oid, err = db.GetOrgIDOf(ctx, id); err != nil {
			return nil, fmt.Errorf("get org: %w", err)
		} else if oid.IsValid() {
			oidsSet[oid.String()] = true
		}

		// Get the resource project.
		if pid, err = db.GetProjectIDOf(ctx, id); err != nil {
			return nil, fmt.Errorf("get project: %w", err)
		} else if pid.IsValid() {
			pidsSet[pid.String()] = true
		}
	}

	associations := make(map[string]map[string]string)

	for name, id := range cfg.associations {
		// In case the resource is associated with a something, we need to get that thing's org.
		// This is relevant, for example, with new builds or sessions, where they are explicitly owned
		// but can be associated with a project. The policy needs to decide if to allow it.

		if id == nil || !id.IsValid() {
			continue
		}

		oid, err := db.GetOrgIDOf(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get project org: %w", err)
		}

		associations[name] = map[string]string{}

		if oid.IsValid() {
			oidsSet[oid.String()] = true
			associations[name]["org_id"] = oid.String()
		}

		if id.Kind() == sdktypes.ProjectIDKind {
			pidsSet[id.String()] = true
			associations[name]["project_id"] = id.String()
		} else {
			pid, err := db.GetProjectIDOf(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("get project id: %w", err)
			}

			if pid.IsValid() {
				pidsSet[pid.String()] = true
				associations[name]["project_id"] = pid.String()
			}
		}

	}

	data := map[string]any{
		"kind":                   id.Kind(),                          // resource kind. available even if id is invalid, as it still contains the kind.
		"user_id":                uid.String(),                       // requester user id if valid, else "".
		"user_org_ids":           uoids,                              // orgs the user is part of.
		"action_type":            actType,                            // [type:]xxx of action, or "" if not specified.
		"action":                 act,                                // [xxx:]action of action.
		"resource_id":            id.String(),                        // resource id.
		"resource_org_id":        oid.String(),                       // resource org id, if resource id is valid. else "".
		"resource_project_id":    pid.String(),                       // resource project tid, if resource id is valid. else "".
		"data":                   cfg.data,                           // aux data supplied by the caller.
		"associated_org_ids":     slices.Collect(maps.Keys(oidsSet)), // all unique non-zero associated org ids and the resource org id.
		"associated_project_ids": slices.Collect(maps.Keys(pidsSet)), // all unique non-zero associated project ids and the resource project id.
		"associations":           associations,                       // name -> {"org_id": "xxx"} of associated resources.
	}

	return data, nil
}
