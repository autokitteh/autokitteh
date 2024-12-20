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

func (f *dbFixture) createProjectConnection(t *testing.T) (scheme.Project, scheme.Connection) {
	p := f.newProject()
	c := f.newConnection(p)

	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c)

	return p, c
}

func preTriggerTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers
	return f
}

func TestCreateTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	tr := f.newTrigger(p, c)

	// test createTrigger
	f.createTriggersAndAssert(t, tr)
}

func TestGetTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	t1 := f.newTrigger(p, c)
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
	p, c := f.createProjectConnection(t)
	f.saveBuildsAndAssert(t, b)

	// negative test with non-existing assets

	// zero ProjectID and ConnectionID
	t1 := f.newTrigger()
	assert.Equal(t, t1.ProjectID, sdktypes.UUID{})
	assert.Nil(t, t1.ConnectionID)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t1), gorm.ErrForeignKeyViolated)

	// change connection to builtin scheduler connection (now user owned) - should be allowed

	// use buildID (owned by user to pass user ownership test) to fake unexisting IDs
	t2 := f.newTrigger(p, c)
	t2.ProjectID = b.BuildID // no such projectID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), gorm.ErrForeignKeyViolated)

	t2.ProjectID = p.ProjectID
	t2.ConnectionID = &b.BuildID // no such connectionID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &t2), gorm.ErrForeignKeyViolated)
	t2.ConnectionID = &c.ConnectionID

	// test with existing assets
	f.createTriggersAndAssert(t, t2)
}

func TestDeleteTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t)
	p, c := f.createProjectConnection(t)

	trg := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, trg)
	evt := f.newEvent(trg, p)
	f.createEventsAndAssert(t, evt)

	// trigger could be deleted, even if it refenced by non-deleted event
	findAndAssertOne(t, f, evt, "trigger_id = ?", trg.TriggerID)    // non-deleted event
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, trg.TriggerID)) // deleted trigger
}

func TestListTriggers(t *testing.T) {
	f := preTriggerTest(t)

	p, c := f.createProjectConnection(t)
	t1 := f.newTrigger(p, c)
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

	p, c := f.createProjectConnection(t)
	tr := f.newTrigger(p, c)
	f.createTriggersAndAssert(t, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	f.assertTriggersDeleted(t, tr)
}

func TestDuplicatedTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p1 := f.newProject()
	p2 := f.newProject()
	c := f.newConnection(p1)
	f.createProjectsAndAssert(t, p1, p2)
	f.createConnectionsAndAssert(t, c)

	// test create two triggers with the same name for the same project and same environment
	tr1 := f.newTrigger(p1, c, "trg")
	f.createTriggersAndAssert(t, tr1)

	tr2 := f.newTrigger(p1, c, "trg")
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr2), gorm.ErrDuplicatedKey)

	// could create the same named trigger for the different project
	tr3 := f.newTrigger(p2, c, "trg")
	f.createTriggersAndAssert(t, tr3)
}
