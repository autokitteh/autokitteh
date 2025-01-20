package authz

import (
	"testing"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbtest"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

var (
	zumi       = sdktypes.NewUser().WithStatus(sdktypes.UserStatusActive).WithEmail("zumi@cats").WithNewID()
	gizmo      = sdktypes.NewUser().WithStatus(sdktypes.UserStatusActive).WithEmail("gizmo@cats").WithNewID()
	sufi       = sdktypes.NewUser().WithStatus(sdktypes.UserStatusActive).WithEmail("sufi@cats").WithNewID()
	shoogy     = sdktypes.NewUser().WithStatus(sdktypes.UserStatusActive).WithEmail("shoogy@dogs").WithNewID()
	cats       = sdktypes.NewOrg().WithNewID()
	zumiInCats = sdktypes.NewOrgMember(cats.ID(), zumi.ID()).WithStatus(sdktypes.OrgMemberStatusActive).WithRoles(sdktypes.NewSymbol("admin"))
	sufiInCats = sdktypes.NewOrgMember(cats.ID(), sufi.ID()).WithStatus(sdktypes.OrgMemberStatusInvited)
	p          = sdktypes.NewProject().WithNewID().WithName(sdktypes.NewSymbol("project")).WithOrgID(cats.ID())
	tr         = sdktypes.NewTrigger(sdktypes.NewSymbol("trigger")).WithWebhook().WithNewID().WithProjectID(p.ID())
)

func setupDB(t *testing.T) db.DB {
	return dbtest.NewTestDB(t, zumi, gizmo, shoogy, sufi, cats, zumiInCats, sufiInCats, p, tr)
}
