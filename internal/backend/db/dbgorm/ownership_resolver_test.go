package dbgorm

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.autokitteh.dev/autokitteh/internal/backend/builds"
	"go.autokitteh.dev/autokitteh/internal/backend/connections"
	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/deployments"
	"go.autokitteh.dev/autokitteh/internal/backend/envs"
	"go.autokitteh.dev/autokitteh/internal/backend/events"
	"go.autokitteh.dev/autokitteh/internal/backend/projects"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions"
	"go.autokitteh.dev/autokitteh/internal/backend/sessions/sessionsvcs"
	"go.autokitteh.dev/autokitteh/internal/backend/triggers"
	"go.autokitteh.dev/autokitteh/internal/backend/vars"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/internal/resolver"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkintegrations"
	"go.autokitteh.dev/autokitteh/sdk/sdkmodule"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.uber.org/zap/zaptest"
)

type dbs struct {
	intSvc sdkservices.Integrations
	prjSvc sdkservices.Projects
	bldSvc sdkservices.Builds
	depSvc sdkservices.Deployments
	envSvc sdkservices.Envs
	conSvc sdkservices.Connections
	sesSvc sdkservices.Sessions
	evtSvc sdkservices.Events
	trgSvc sdkservices.Triggers
	varSvc sdkservices.Vars
}

// type integrations struct {
// 	db db.DB
// 	z  *zap.Logger
// }
// func newIntegrationSvc(db db.DB, z *zap.Logger) sdkservices.Integrations {
// 	return &integrations{db: db, z: z}
// }

// func (isvc *integrations) GetByID(ctx context.Context, id sdktypes.IntegrationID) (sdktypes.Integration, error) {
// 	return isvc.db.GetIntegration(ctx, id)
// }

// func (isvc *integrations) GetByName(ctx context.Context, name sdktypes.Symbol) (sdktypes.Integration, error) {
// 	ii, err := isvc.db.ListIntegrations(ctx)
// 	if err == nil {
// 		for _, i := range ii {
// 			if i.DisplayName() == name.String() {
// 				return i, nil
// 			}
// 		}
// 	}
// 	return sdktypes.InvalidIntegration, err
// }

// func (isvc *integrations) List(ctx context.Context, nameSubstring string) ([]sdktypes.Integration, error) {
// 	ii, err := isvc.db.ListIntegrations(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var iiFiltered []sdktypes.Integration

//		for _, i := range ii {
//			if strings.Contains(i.DisplayName(), nameSubstring) {
//				iiFiltered = append(iiFiltered, i)
//			}
//		}
//		return iiFiltered, nil
//	}
// type ConnectionsSvcWithoutEnrich struct {
// 	connections.Connections
// }

// func newConnectionsSvc(db db.DB, z *zap.Logger) sdkservices.Connections {
// 	return &ConnectionsSvcWithoutEnrich{connections.Connections{Z: z, DB: db, Name: "aaa"}}
// }

// func (c *ConnectionsSvcWithoutEnrich) enrichConnection(_ context.Context, conn sdktypes.Connection) (sdktypes.Connection, error) {
// 	return conn, nil
// }

func newIntegrationsSvc() sdkservices.Integrations {
	integrationID := sdktypes.NewIntegrationIDFromName("test")

	desc := kittehs.Must1(sdktypes.StrictIntegrationFromProto(&sdktypes.IntegrationPB{
		IntegrationId: integrationID.String(),
		UniqueName:    "test",
		DisplayName:   "Test",
		Description:   "Test integration",
	}))

	testIntegration := sdkintegrations.NewIntegration(desc, sdkmodule.New())
	return sdkintegrations.New([]sdkservices.Integration{testIntegration})
}

func (dbs *dbs) Builds() sdkservices.Builds             { return dbs.bldSvc }
func (dbs *dbs) Deployments() sdkservices.Deployments   { return dbs.depSvc }
func (dbs *dbs) Projects() sdkservices.Projects         { return dbs.prjSvc }
func (dbs *dbs) Envs() sdkservices.Envs                 { return dbs.envSvc }
func (dbs *dbs) Connections() sdkservices.Connections   { return dbs.conSvc }
func (dbs *dbs) Sessions() sdkservices.Sessions         { return dbs.sesSvc }
func (dbs *dbs) Events() sdkservices.Events             { return dbs.evtSvc }
func (dbs *dbs) Triggers() sdkservices.Triggers         { return dbs.trgSvc }
func (dbs *dbs) Vars() sdkservices.Vars                 { return dbs.varSvc }
func (dbs *dbs) Integrations() sdkservices.Integrations { return dbs.intSvc }

func newDBServices(t *testing.T) (sdkservices.DBServices, *dbFixture) {
	f := newDBFixture().WithDebug()
	var gdb db.DB = f.gormdb

	z := zaptest.NewLogger(t) // FIXME: or gormdb.z?

	intSvc := newIntegrationsSvc()
	bldSvc := builds.New(builds.Builds{Z: z, DB: gdb})
	prjSvc := projects.New(projects.Projects{Z: z, DB: gdb})
	depSvc := deployments.New(z, gdb)
	conSvc := connections.New(connections.Connections{Z: z, DB: gdb, Integrations: intSvc})
	envSvc := envs.New(z, gdb)
	evtSvc := events.New(z, gdb)
	trgSvc := triggers.New(z, gdb, nil)
	varSvc := vars.New(z, gdb, nil)
	sesSvc := sessions.New(z, nil, gdb,
		sessionsvcs.Svcs{DB: gdb, Builds: bldSvc, Connections: conSvc, Deployments: depSvc, Envs: envSvc, Triggers: trgSvc, Vars: varSvc})

	return &dbs{
		intSvc: intSvc,
		prjSvc: prjSvc,
		bldSvc: bldSvc,
		depSvc: depSvc,
		envSvc: envSvc,
		conSvc: conSvc,
		sesSvc: sesSvc,
		evtSvc: evtSvc,
		trgSvc: trgSvc,
		varSvc: varSvc,
	}, f
}

func createResolverAndFixture(t *testing.T) (resolver.Resolver, *dbFixture) {
	dbServices, f := newDBServices(t)
	r := resolver.Resolver{Client: dbServices}
	return r, f
}

func TestResolverBuildIDWithOwnership(t *testing.T) {
	r, f := createResolverAndFixture(t)
	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)

	bid := sdktypes.NewIDFromUUID[sdktypes.BuildID](&b.BuildID)
	bids := bid.String()

	// resolve ok
	b1, _, err := r.BuildID(f.ctx, bids)
	assert.NoError(t, err)
	assert.Equal(t, bid, b1.ID())

	// fail due to auth
	f.ctx = withUser(f.ctx, u)
	_, _, err = r.BuildID(f.ctx, bids)
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
}

func TestResolverDeploymentIDWithOwnership(t *testing.T) {
	r, f := createResolverAndFixture(t)

	_, b, e := f.createProjectBuildEnv(t)
	d := f.newDeployment(b, e)
	f.createDeploymentsAndAssert(t, d)

	did := sdktypes.NewIDFromUUID[sdktypes.DeploymentID](&d.DeploymentID)
	dids := did.String()

	// resolve ok
	d1, _, err := r.DeploymentID(f.ctx, dids)
	assert.NoError(t, err)
	assert.Equal(t, did, d1.ID())

	// fail due to auth
	f.ctx = withUser(f.ctx, u)
	_, _, err = r.DeploymentID(f.ctx, dids)
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
}

func TestResolverEventIDWithOwnership(t *testing.T) {
	r, f := createResolverAndFixture(t)

	c := f.newConnection()
	e := f.newEvent(c)
	f.createConnectionsAndAssert(t, c)
	f.createEventsAndAssert(t, e)

	eid := sdktypes.NewIDFromUUID[sdktypes.EventID](&e.EventID)
	eids := eid.String()

	// resolve ok
	e1, _, err := r.EventID(f.ctx, eids)
	assert.NoError(t, err)
	assert.Equal(t, eid, e1.ID())

	// fail due to auth
	f.ctx = withUser(f.ctx, u)
	_, _, err = r.EventID(f.ctx, eids)
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
}

func TestResolverTriggerIDWithOwnership(t *testing.T) {
	r, f := createResolverAndFixture(t)

	p, c, e := f.createProjectConnectionEnv(t)
	trg := f.newTrigger(p, e, c)

	f.createTriggersAndAssert(t, trg)

	tid := sdktypes.NewIDFromUUID[sdktypes.TriggerID](&trg.TriggerID)
	tids := tid.String()

	// resolve ok
	t1, _, err := r.TriggerID(f.ctx, tids)
	assert.NoError(t, err)
	assert.Equal(t, tid, t1.ID())

	// fail due to auth
	f.ctx = withUser(f.ctx, u)
	_, _, err = r.TriggerID(f.ctx, tids)
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
}

func TestResolverSessionIDWithOwnership(t *testing.T) {
	r, f := createResolverAndFixture(t)

	b := f.newBuild()
	s := f.newSession(b)
	f.saveBuildsAndAssert(t, b)
	f.createSessionsAndAssert(t, s)

	sid := sdktypes.NewIDFromUUID[sdktypes.SessionID](&s.SessionID)
	sids := sid.String()

	// resolve ok
	s1, _, err := r.SessionID(f.ctx, sids)
	assert.NoError(t, err)
	assert.Equal(t, sid, s1.ID())

	// fail due to auth
	f.ctx = withUser(f.ctx, u)
	_, _, err = r.SessionID(f.ctx, sids)
	assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
}

func TestResolverConnectionNameOrIdWithOwnership(t *testing.T) {
	r, f := createResolverAndFixture(t)

	// since we are passing via service, connection should be defined properly otherwise will fail in parsing
	// e.g. have project and integration defined

	i := f.newIntegration()
	i.IntegrationID = sdktypes.NewIntegrationIDFromName("test").UUIDValue()
	p := f.newProject()
	c := f.newConnection(p, i)
	f.createIntegrationsAndAssert(t, i)
	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c)

	cid := sdktypes.NewIDFromUUID[sdktypes.ConnectionID](&c.ConnectionID)
	cids := cid.String()

	testCases := []struct {
		name     string
		nameOrID string
		project  string
	}{
		{"connectionID", cids, ""},
		{"connectionName,projectName", c.Name, p.Name},
		{"projectName/connectionName", fmt.Sprintf("%s/%s", p.Name, c.Name), ""},
	}

	// resolve ok
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c, _, err := r.ConnectionNameOrID(f.ctx, tc.nameOrID, tc.project)
			assert.NoError(t, err)
			assert.Equal(t, cid, c.ID())
		})
	}

	// fail due to auth
	f.ctx = withUser(f.ctx, u)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := r.ConnectionNameOrID(f.ctx, tc.nameOrID, tc.project)
			assert.ErrorIs(t, err, sdkerrors.ErrNotFound)
		})
	}

	// TODO: check that all users could access default cronConnection?
}

// FIXME:
// EnvNameOrID
// IntegraitonNameOrID
// ProjectNameOrID
