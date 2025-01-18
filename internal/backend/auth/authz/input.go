package authz

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func hydrate(ctx context.Context, db db.DB, id sdktypes.ID, obj sdktypes.Object) (map[string]any, error) {
	if id == nil {
		return nil, errors.New("hydrate: id is nil")
	}

	m := map[string]any{"kind": id.Kind()}

	if !id.IsValid() {
		return m, nil
	}

	m["id"] = id.String()

	switch id.Kind() {
	case sdktypes.UserIDKind:
		ms, _, err := db.GetOrgsForUser(ctx, id.(sdktypes.UserID), false)
		if err != nil {
			return nil, fmt.Errorf("get orgs for user: %w", err)
		}

		m["org_memberships"] = kittehs.ListToMap(ms, func(m sdktypes.OrgMember) (string, any) {
			return m.OrgID().String(), map[string]any{
				"status": m.Status().String(),
				"roles":  kittehs.TransformToStrings(m.Roles()),
			}
		})

		var u sdktypes.User
		if obj != nil && obj.IsValid() {
			u = obj.(sdktypes.User)
		} else if u, err = db.GetUser(ctx, id.(sdktypes.UserID), ""); err != nil {
			return nil, fmt.Errorf("get user: %w", err)
		}

		m["email"] = u.Email()
		m["status"] = u.Status().String()

	case sdktypes.OrgIDKind:
		m["org_id"] = id.String()

	case sdktypes.ProjectIDKind:
		oid, err := db.GetOrgIDOf(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get org: %w", err)
		}

		m["org_id"] = oid.String()

	default:
		oid, err := db.GetOrgIDOf(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get org: %w", err)
		}

		pid, err := db.GetProjectIDOf(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("get project: %w", err)
		}

		if pid.IsValid() {
			m["project_id"] = pid.String()
		}

		if oid.IsValid() {
			m["org_id"] = oid.String()
		}
	}

	return m, nil
}

func buildInput(ctx context.Context, db db.DB, id sdktypes.ID, action string, cfg checkCfg) (map[string]any, error) {
	authnUser := authcontext.GetAuthnUser(ctx)
	if !authnUser.IsValid() {
		return nil, sdkerrors.ErrUnauthenticated
	}

	m, err := hydrate(ctx, db, authnUser.ID(), authnUser)
	if err != nil {
		return nil, err
	}

	rsc, err := hydrate(ctx, db, id, nil)
	if err != nil {
		return nil, err
	}

	associations := make(map[string]map[string]any, len(cfg.associations)+1)

	associations["subject"] = rsc

	for name, id := range cfg.associations {
		// In case the subject is associated with a something, we need to get that thing's org.
		// This is relevant, for example, with new builds or sessions, where they are explicitly owned
		// but can be associated with a project. The policy needs to decide if to allow it.

		if id == nil || !id.IsValid() {
			continue
		}

		if associations[name], err = hydrate(ctx, db, id, nil); err != nil {
			return nil, err
		}
	}

	// TODO: Use https://docs.styra.com/opa/rego-by-example/builtins/regex/globs_match to match permissions.
	actType, act, ok := strings.Cut(action, ":")
	if !ok {
		actType, act = "", action
	}

	return map[string]any{
		"action": map[string]any{
			"full": action,  // full action as supplied by the caller.
			"type": actType, // [xxx:]action of full.
			"name": act,     // [type:]xxx of full, or "" if not specified.
		},

		"data":         cfg.data,     // aux data supplied by the caller.
		"authn_user":   m,            // hydrated authenticated user.
		"subject":      rsc,          // hydrated subject.
		"associations": associations, // hydrated associations.
	}, nil
}
