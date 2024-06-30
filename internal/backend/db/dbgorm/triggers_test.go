package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func (f *dbFixture) createTriggersAndAssert(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assert.NoError(t, f.gormdb.createTrigger(f.ctx, &trigger))
		findAndAssertOne(t, f, trigger, "trigger_id = ?", trigger.TriggerID)
	}
}

func (f *dbFixture) assertTriggersDeleted(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assertDeleted(t, f, trigger)
	}
}

func createProjectConnectionEnv(t *testing.T, f *dbFixture) (scheme.Project, scheme.Connection, scheme.Env) {
	p := f.newProject()
	c := f.newConnection(p)
	env := f.newEnv(p)

	f.createProjectsAndAssert(t, p)
	f.createConnectionsAndAssert(t, c)
	f.createEnvsAndAssert(t, env)

	return p, c, env
}

func preTriggerTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers
	return f
}

func TestCreateTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c, env := createProjectConnectionEnv(t, f)
	tr := f.newTrigger(p, c, env)

	// test createTrigger
	f.createTriggersAndAssert(t, tr)
}

func TestGetTrigger(t *testing.T) {
	f := preTriggerTest(t)

	p, c, env := createProjectConnectionEnv(t, f)
	t1 := f.newTrigger(p, c, env)
	f.createTriggersAndAssert(t, t1)

	// test getTrigger
	t2, err := f.gormdb.getTrigger(f.ctx, t1.TriggerID)
	assert.NoError(t, err)
	assert.Equal(t, t1, *t2)

	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, t1.TriggerID))
	_, err = f.gormdb.getTrigger(f.ctx, t1.TriggerID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestCreateTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t)

	b := f.newBuild()
	f.saveBuildsAndAssert(t, b)
	p, c, env := createProjectConnectionEnv(t, f)
	tr := f.newTrigger(p, c, env)

	// negative test with non-existing assets
	// use buildID (owned by user to pass user ownership test) as unexisting ID

	tr.EnvID = b.BuildID // no such envID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr), gorm.ErrForeignKeyViolated)
	tr.EnvID = env.EnvID

	tr.ConnectionID = b.BuildID // no such connectionID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr), gorm.ErrForeignKeyViolated)
	tr.ConnectionID = c.ConnectionID

	// test with existing assets
	f.createTriggersAndAssert(t, tr)
}

func TestListTriggers(t *testing.T) {
	f := preTriggerTest(t)

	p, c, env := createProjectConnectionEnv(t, f)
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

	p, c, env := createProjectConnectionEnv(t, f)
	tr := f.newTrigger(p, c, env)
	f.createTriggersAndAssert(t, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	f.assertTriggersDeleted(t, tr)
}
