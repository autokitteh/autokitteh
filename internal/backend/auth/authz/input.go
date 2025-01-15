package authz

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func buildInput(ctx context.Context, db db.DB, id sdktypes.ID, action string, cfg checkCfg) (map[string]any, error) {
	authnUser := authcontext.GetAuthnUser(ctx)
	if !authnUser.IsValid() {
		return nil, sdkerrors.ErrUnauthenticated
	}

	memberships, err := db.GetOrgsForUser(ctx, authnUser.ID())
	if err != nil {
		return nil, fmt.Errorf("get orgs for user: %w", err)
	}

	membershipsMap := kittehs.ListToMap(memberships, func(m sdktypes.OrgMember) (string, any) {
		return m.OrgID().String(), map[string]any{
			"status": m.Status().String(),
			"roles":  kittehs.TransformToStrings(m.Roles()),
		}
	})

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

	associations := make(map[string]map[string]any)

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

		associations[name] = map[string]any{
			"id": id.String(),
		}

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
		// Authenticated user (requester) information.
		"authn_user":      authnUser,               // requester authn user.
		"authn_user_id":   authnUser.ID().String(), // requester authn user id.
		"authn_user_orgs": membershipsMap,          // org memberships related to the authn user: {org_id: {"status": "INVITED", "roles": [str]}}

		// Resource information.
		"kind":                id.Kind(),    // resource kind. available even if id is invalid, as it still contains the kind.
		"action_type":         actType,      // [type:]xxx of action, or "" if not specified.
		"action":              act,          // [xxx:]action of action.
		"resource_id":         id.String(),  // resource id.
		"resource_org_id":     oid.String(), // resource org id, if resource id is valid. else "".
		"resource_project_id": pid.String(), // resource project id, if resource id is valid. else "".
		"data":                cfg.data,     // aux data supplied by the caller.

		// Explicit associations given as part of the check call.
		"associated_org_ids":     slices.Collect(maps.Keys(oidsSet)), // all unique non-zero associated org ids and the resource org id.
		"associated_project_ids": slices.Collect(maps.Keys(pidsSet)), // all unique non-zero associated project ids and the resource project id.
		"associations":           associations,                       // name -> {"org_id": "xxx"} of associated resources.
	}

	return data, nil
}
