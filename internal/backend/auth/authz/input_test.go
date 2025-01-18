package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestBuildInputUser(t *testing.T) {
	db := setupDB(t)

	tests := []struct {
		name     string
		authn    sdktypes.User // authenticated user
		id       sdktypes.ID   // subject id
		action   string
		opts     []CheckOpt
		expected map[string]any
	}{
		{
			name:   "empty id",
			authn:  gizmo,
			id:     sdktypes.InvalidProjectID,
			action: "action_type:action",
			opts: []CheckOpt{
				WithData("project", p),
			},
			expected: map[string]any{
				"action": map[string]any{
					"name": "action",
					"type": "action_type",
					"full": "action_type:action",
				},
				"associations": map[string]map[string]any{
					"subject": {
						"kind": "prj",
					},
				},
				"data": map[string]any{
					"project": p,
				},
				"subject": map[string]any{
					"kind": "prj",
				},
				"authn_user": map[string]any{
					"id":              gizmo.ID().String(),
					"kind":            "usr",
					"email":           "gizmo@cats",
					"org_memberships": map[string]any{},
					"status":          "ACTIVE",
				},
			},
		},
		{
			name:   "associations",
			authn:  gizmo,
			id:     p.ID(),
			action: "action_type:action",
			opts: []CheckOpt{
				WithAssociationWithID("project", p.ID()),
				WithAssociationWithID("user", gizmo.ID()),
				WithAssociationWithID("org", cats.ID()),
			},
			expected: map[string]any{
				"action": map[string]any{
					"name": "action",
					"type": "action_type",
					"full": "action_type:action",
				},
				"associations": map[string]map[string]any{
					"subject": {
						"kind":   "prj",
						"id":     p.ID().String(),
						"org_id": cats.ID().String(),
					},
					"user": {
						"id":              gizmo.ID().String(),
						"kind":            "usr",
						"email":           "gizmo@cats",
						"org_memberships": map[string]any{},
						"status":          "ACTIVE",
					},
					"project": {
						"id":     p.ID().String(),
						"org_id": cats.ID().String(),
						"kind":   "prj",
					},
					"org": {
						"id":     cats.ID().String(),
						"org_id": cats.ID().String(),
						"kind":   "org",
					},
				},
				"data": map[string]any{},
				"subject": map[string]any{
					"kind":   "prj",
					"id":     p.ID().String(),
					"org_id": cats.ID().String(),
				},
				"authn_user": map[string]any{
					"id":              gizmo.ID().String(),
					"kind":            "usr",
					"email":           "gizmo@cats",
					"org_memberships": map[string]any{},
					"status":          "ACTIVE",
				},
			},
		},
		{
			name:   "project",
			authn:  gizmo,
			id:     p.ID(),
			action: "action_type:action",
			expected: map[string]any{
				"action": map[string]any{
					"name": "action",
					"type": "action_type",
					"full": "action_type:action",
				},
				"associations": map[string]map[string]any{
					"subject": {
						"kind":   "prj",
						"id":     p.ID().String(),
						"org_id": cats.ID().String(),
					},
				},
				"data": map[string]any{},
				"subject": map[string]any{
					"kind":   "prj",
					"id":     p.ID().String(),
					"org_id": cats.ID().String(),
				},
				"authn_user": map[string]any{
					"id":              gizmo.ID().String(),
					"kind":            "usr",
					"email":           "gizmo@cats",
					"org_memberships": map[string]any{},
					"status":          "ACTIVE",
				},
			},
		},
		{
			name:   "user",
			authn:  zumi,
			id:     gizmo.ID(),
			action: "action_type:action",
			expected: map[string]any{
				"action": map[string]any{
					"name": "action",
					"type": "action_type",
					"full": "action_type:action",
				},
				"data": map[string]any{},
				"subject": map[string]any{
					"kind":            "usr",
					"id":              gizmo.ID().String(),
					"email":           "gizmo@cats",
					"org_memberships": map[string]any{},
					"status":          "ACTIVE",
				},
				"associations": map[string]map[string]any{
					"subject": {
						"kind":            "usr",
						"id":              gizmo.ID().String(),
						"email":           "gizmo@cats",
						"org_memberships": map[string]any{},
						"status":          "ACTIVE",
					},
				},
				"authn_user": map[string]any{
					"id":    zumi.ID().String(),
					"kind":  "usr",
					"email": "zumi@cats",
					"org_memberships": map[string]any{
						cats.ID().String(): map[string]any{
							"roles":  []string{"admin"},
							"status": "ACTIVE",
						},
					},
					"status": "ACTIVE",
				},
			},
		},
		{
			name:   "trigger",
			authn:  gizmo,
			id:     tr.ID(),
			action: "action_type:action",
			opts: []CheckOpt{
				WithData("cat", "meow"),
			},
			expected: map[string]any{
				"action": map[string]any{
					"name": "action",
					"type": "action_type",
					"full": "action_type:action",
				},
				"associations": map[string]map[string]any{
					"subject": {
						"kind":       "trg",
						"id":         tr.ID().String(),
						"org_id":     cats.ID().String(),
						"project_id": p.ID().String(),
					},
				},
				"subject": map[string]any{
					"kind":       "trg",
					"id":         tr.ID().String(),
					"org_id":     cats.ID().String(),
					"project_id": p.ID().String(),
				},
				"data": map[string]any{
					"cat": "meow",
				},
				"authn_user": map[string]any{
					"id":              gizmo.ID().String(),
					"kind":            "usr",
					"email":           "gizmo@cats",
					"org_memberships": map[string]any{},
					"status":          "ACTIVE",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			if test.authn.IsValid() {
				ctx = authcontext.SetAuthnUser(context.Background(), test.authn)
			}

			inputs, err := buildInput(ctx, db, test.id, test.action, configure(test.opts))
			if assert.NoError(t, err) {
				assert.Equal(t, test.expected, inputs)
			}
		})
	}
}
