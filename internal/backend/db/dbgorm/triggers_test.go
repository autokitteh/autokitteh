package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (f *dbFixture) createTriggersAndAssert(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assert.NoError(t, f.gormdb.createTrigger(f.ctx, &trigger))
		findAndAssertOne(t, f, trigger, "trigger_id = ?", trigger.TriggerID)
	}
}

func (f *dbFixture) assertTriggersDeleted(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assertSoftDeleted(t, f, trigger)
	}
}

func (f *dbFixture) createProjectConnectionEnv(t *testing.T) (scheme.Project, scheme.Connection, scheme.Env) {
	p := f.newProject()
	c := f.newConnection(p)
	env := f.newEnv(p)

	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c)
	f.createEnvsAndAssert(t, env)

	return p, c, env
}

func preTriggerTest(t *testing.T) *dbFixture {
	f := newDBFixture().withUser(sdktypes.DefaultUser)
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers
	return f
}

func TestCreateTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c, env := f.createProjectConnectionEnv(t)
	tr := f.newTrigger(p, c, env)

	// test createTrigger
	f.createTriggersAndAssert(t, tr)
}

func TestGetTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c, env := f.createProjectConnectionEnv(t)
	t1 := f.newTrigger(p, c, env)
	f.createTriggersAndAssert(t, t1)

	// test getTrigger
	t2, err := f.gormdb.getTriggerByID(f.ctx, t1.TriggerID)
	assert.NoError(t, err)
	assert.Equal(t, t1, *t2)

	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, t1.TriggerID))
	_, err = f.gormdb.getTriggerByID(f.ctx, t1.TriggerID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t)

	b := f.newBuild()
	p, c, e := f.createProjectConnectionEnv(t)
	f.saveBuildsAndAssert(t, b)

	// negative test with non-existing assets

	// zero ProjectID, EnvID and ConnectionID
	t1 := f.newTrigger()
	assert.Equal(t, t1.ProjectID, sdktypes.UUID{})
	assert.Equal(t, t1.EnvID, sdktypes.UUID{})
	assert.Nil(t, t1.ConnectionID)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t1), gorm.ErrForeignKeyViolated)

	// change connection to builtin scheduler connection (now user owned) - should be allowed

	// use buildID (owned by user to pass user ownership test) to fake unexisting IDs
	t2 := f.newTrigger(p, c, e)
	t2.EnvID = b.BuildID // no such envID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), gorm.ErrForeignKeyViolated)
	t2.EnvID = e.EnvID

	t2.ConnectionID = &b.BuildID // no such connectionID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), gorm.ErrForeignKeyViolated)
	t2.ConnectionID = &c.ConnectionID

	// test with existing assets
	f.createTriggersAndAssert(t, t2)
}

func TestDeleteTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t).WithDebug()
	p, c, e := f.createProjectConnectionEnv(t)

	trg := f.newTrigger(p, c, e)
	f.createTriggersAndAssert(t, trg)
	evt := f.newEvent(trg)
	f.createEventsAndAssert(t, evt)

	// trigger could be deleted, even if it refenced by event, since trigger is soft deleted
	assert.ErrorIs(t, f.gormdb.deleteTrigger(f.ctx, trg.TriggerID), gorm.ErrForeignKeyViolated)

	_ = f.gormdb.deleteEvent(f.ctx, evt.EventID)
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, trg.TriggerID))
}

func TestListTriggers(t *testing.T) {
	f := preTriggerTest(t)

	p, c, env := f.createProjectConnectionEnv(t)
	t1 := f.newTrigger(p, c, env)
	f.createTriggersAndAssert(t, t1)

	// test listTriggers
	triggers, err := f.gormdb.listTriggers(f.ctx, sdkservices.ListTriggersFilter{})
	assert.NoError(t, err)
	assert.Len(t, triggers, 1)
	assert.Equal(t, t1, triggers[0])

	// test listTriggers after delete
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, t1.TriggerID))
	triggers, err = f.gormdb.listTriggers(f.ctx, sdkservices.ListTriggersFilter{})
	assert.NoError(t, err)
	assert.Len(t, triggers, 0)
}

func TestDeleteTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c, env := f.createProjectConnectionEnv(t)
	tr := f.newTrigger(p, c, env)
	f.createTriggersAndAssert(t, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	f.assertTriggersDeleted(t, tr)
}

func TestDuplicatedTrigger(t *testing.T) {
	f := preTriggerTest(t).WithDebug()

	p1 := f.newProject()
	p2 := f.newProject()
	c := f.newConnection() // trigger doesn't check for connection's projectID
	e1 := f.newEnv(p1, "env")
	e11 := f.newEnv(p1, "env1")
	e2 := f.newEnv(p2, "env")
	f.createProjectsAndAssert(t, p1, p2)
	f.createConnectionsAndAssert(t, c)
	f.createEnvsAndAssert(t, e1, e11, e2)

	// test create two triggers with the same name for the same project and same environment
	tr1 := f.newTrigger(p1, c, e1, "trg")
	f.createTriggersAndAssert(t, tr1)

	tr2 := f.newTrigger(p1, c, e1, "trg")
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr2), gorm.ErrDuplicatedKey)

	// could create the same named trigger for the same project but different environment
	tr3 := f.newTrigger(p1, c, e11, "trg")
	f.createTriggersAndAssert(t, tr3)

	// could create the same named trigger for the different project
	tr4 := f.newTrigger(p2, c, e2, "trg")
	f.createTriggersAndAssert(t, tr4)
}
