package dbgorm_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func TestGetOwnerOrgID(t *testing.T) {
	bid := sdktypes.NewBuildID()
	cid := sdktypes.NewConnectionID()
	did := sdktypes.NewDeploymentID()
	oid := sdktypes.NewOrgID()
	pid := sdktypes.NewProjectID()
	sid := sdktypes.NewSessionID()
	tid := sdktypes.NewTriggerID()

	db, err := dbgorm.New(zap.NewNop(), &dbgorm.Config{})
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, db.Connect(ctx))
	// No need for migrations inside unit tests:
	// require.NoError(t, db.Setup(ctx))

	_, err = db.CreateOrg(ctx, sdktypes.NewOrg().WithID(oid))
	require.NoError(t, err)

	require.NoError(t, db.CreateProject(ctx, sdktypes.NewProject().WithName(sdktypes.NewSymbol("test")).WithID(pid).WithOrgID(oid)))
	require.NoError(t, db.SaveBuild(ctx, sdktypes.NewBuild().WithProjectID(pid).WithID(bid), []byte("meow")))
	require.NoError(t, db.CreateConnection(ctx, sdktypes.NewConnection(cid).WithName(sdktypes.NewSymbol("test")).WithIntegrationID(sdktypes.NewIntegrationID()).WithProjectID(pid)))
	require.NoError(t, db.CreateTrigger(ctx, sdktypes.NewTrigger(sdktypes.NewSymbol("test")).WithProjectID(pid).WithID(tid).WithConnectionID(cid)))
	require.NoError(t, db.CreateDeployment(ctx, sdktypes.NewDeployment(did, pid, bid)))

	loc, err := sdktypes.ParseCodeLocation("foo.py:moo")
	require.NoError(t, err)

	require.NoError(t, db.CreateSession(ctx, sdktypes.NewSession(bid, loc, nil, nil).WithID(sid).WithProjectID(pid)))

	tests := []struct {
		name string
		id   sdktypes.ID
	}{
		{"build", bid},
		{"connection", cid},
		{"connection_scope", sdktypes.NewVarScopeID(cid)},
		{"deployment", did},
		{"project", pid},
		{"project_scope", sdktypes.NewVarScopeID(pid)},
		{"session", sid},
		{"trigger", tid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := db.GetOrgIDOf(ctx, tt.id)
			assert.NoError(t, err)
			assert.Equal(t, oid.String(), got.String())
		})
	}
}
