package dbgorm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestGetOwnerOrgID(t *testing.T) {
	pid := sdktypes.NewProjectID()
	bid := sdktypes.NewBuildID()
	cid := sdktypes.NewConnectionID()
	tid := sdktypes.NewTriggerID()
	did := sdktypes.NewDeploymentID()
	sid := sdktypes.NewSessionID()

	db, err := dbgorm.New(zap.NewNop(), &dbgorm.Config{})
	require.NoError(t, err)

	ctx := context.Background()

	require.NoError(t, db.Connect(ctx))
	require.NoError(t, db.Setup(ctx))

	oid := sdktypes.NewOrgID()

	_, err = db.CreateOrg(ctx, sdktypes.NewOrg().WithID(oid))
	require.NoError(t, err)

	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test")).WithID(pid).WithOrgID(oid)))
	require.NoError(t, db.SaveBuild(ctx, sdktypes.NewBuild().WithProjectID(pid).WithID(bid), []byte("meow")))
	require.NoError(t, db.CreateConnection(ctx, sdktypes.NewConnection(cid).WithName(sdktypes.NewSymbol("test")).WithIntegrationID(sdktypes.NewIntegrationID()).WithProjectID(pid)))
	require.NoError(t, db.CreateTrigger(ctx, sdktypes.NewTrigger(sdktypes.NewSymbol("test")).WithProjectID(pid).WithID(tid).WithConnectionID(cid)))
	require.NoError(t, db.CreateDeployment(ctx, sdktypes.NewDeployment(did, pid, bid)))

	loc := kittehs.Must1(sdktypes.ParseCodeLocation("foo.py:moo"))
	require.NoError(t, db.CreateSession(ctx, sdktypes.NewSession(bid, loc, nil, nil).WithID(sid).WithProjectID(pid)))

	ids := []sdktypes.ID{
		pid,
		bid,
		cid,
		tid,
		did,
		sid,
		sdktypes.NewVarScopeID(pid),
		sdktypes.NewVarScopeID(cid),
	}

	for _, id := range ids {
		t.Run(id.String(), func(t *testing.T) {
			toid, err := db.GetOrgIDOf(ctx, id)
			if assert.NoError(t, err) {
				assert.Equal(t, oid.String(), toid.String())
			}
		})
	}
}
