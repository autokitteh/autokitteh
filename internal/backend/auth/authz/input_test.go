package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbtest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestBuildInputUser(t *testing.T) {
	u1 := sdktypes.NewUser().WithStatus(sdktypes.UserStatusActive).WithEmail("zumi@cats").WithNewID()
	u2 := sdktypes.NewUser().WithStatus(sdktypes.UserStatusActive).WithEmail("gizmo@cats").WithNewID()
	o := sdktypes.NewOrg().WithNewID()
	m := sdktypes.NewOrgMember(o.ID(), u1.ID()).WithStatus(sdktypes.OrgMemberStatusActive).WithRoles(sdktypes.NewSymbol("admin"))
	p := sdktypes.NewProject().WithNewID().WithName(sdktypes.NewSymbol("project")).WithOrgID(o.ID())
	db := dbtest.NewTestDB(t, u1, u2, o, m, p)

	tests := []struct {
		name     string
		authn    sdktypes.User // authenticated user
		id       sdktypes.ID   // resource id
		action   string
		opts     []CheckOpt
		expected map[string]any
	}{
		{
			name:   "associations",
			authn:  u2,
			id:     p.ID(),
			action: "action_type:action",
			opts: []CheckOpt{
				WithAssociationWithID("project", p.ID()),
				WithAssociationWithID("user", u2.ID()),
				WithAssociationWithID("org", o.ID()),
			},
			expected: map[string]any{
				"action":                 "action",
				"action_type":            "action_type",
				"associated_org_ids":     []string{o.ID().String()},
				"associated_project_ids": []string{p.ID().String()},
				"associations": map[string]map[string]any{
					"user": {
						"id": u2.ID().String(),
					},
					"project": {
						"id":         p.ID().String(),
						"org_id":     o.ID().String(),
						"project_id": p.ID().String(),
					},
					"org": {
						"id":     o.ID().String(),
						"org_id": o.ID().String(),
					},
				},
				"data":                map[string]any(nil),
				"kind":                "prj",
				"resource_id":         p.ID().String(),
				"resource_project_id": p.ID().String(),
				"resource_org_id":     o.ID().String(),
				"authn_user_id":       u2.ID().String(),
				"authn_user":          u2,
				"authn_user_orgs":     map[string]any{},
			},
		},
		{
			name:   "project",
			authn:  u2,
			id:     p.ID(),
			action: "action_type:action",
			expected: map[string]any{
				"action":                 "action",
				"action_type":            "action_type",
				"associated_org_ids":     []string{o.ID().String()},
				"associated_project_ids": []string{p.ID().String()},
				"associations":           map[string]map[string]any{},
				"data":                   map[string]any(nil),
				"kind":                   "prj",
				"resource_id":            p.ID().String(),
				"resource_project_id":    p.ID().String(),
				"resource_org_id":        o.ID().String(),
				"authn_user_id":          u2.ID().String(),
				"authn_user":             u2,
				"authn_user_orgs":        map[string]any{},
			},
		},
		{
			name:   "user",
			authn:  u1,
			id:     u2.ID(),
			action: "action_type:action",
			expected: map[string]any{
				"action":                 "action",
				"action_type":            "action_type",
				"associated_org_ids":     []string(nil),
				"associated_project_ids": []string(nil),
				"associations":           map[string]map[string]any{},
				"data":                   map[string]any(nil),
				"kind":                   "usr",
				"resource_id":            u2.ID().String(),
				"resource_project_id":    "",
				"resource_org_id":        "",
				"authn_user_id":          u1.ID().String(),
				"authn_user":             u1,
				"authn_user_orgs": map[string]any{
					o.ID().String(): map[string]any{
						"roles":  []string{"admin"},
						"status": "ACTIVE",
					},
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

			var cfg checkCfg
			for _, opt := range test.opts {
				opt(&cfg)
			}

			inputs, err := buildInput(ctx, db, test.id, test.action, cfg)
			if assert.NoError(t, err) {
				assert.Equal(t, test.expected, inputs)
			}
		})
	}
}
