package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
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

func preTriggerTest(t *testing.T, no_foreign_keys bool) *dbFixture {
	f := newDBFixtureFK(no_foreign_keys)
	findAndAssertCount(t, f, scheme.Trigger{}, 0, "") // no trigger
	return f
}

func TestCreateTrigger(t *testing.T) {
	f := preTriggerTest(t, true) // no foreign keys

	tr := f.newTrigger()
	// test createTrigger
	f.createTriggersAndAssert(t, tr)
}

func TestCreateTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t, false) // with foreign keys

	// prepare
	tr := f.newTrigger()
	p := f.newProject()
	env := f.newEnv()
	conn := f.newConnection()

	tr.ProjectID = p.ProjectID
	tr.EnvID = env.EnvID
	tr.ConnectionID = conn.ConnectionID

	f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, env)
	f.createConnectionsAndAssert(t, conn)

	// negative test with non-existing assets
	unexisting := "unexisting"

	tr.ProjectID = unexisting
	assert.ErrorContains(t, f.gormdb.createTrigger(f.ctx, &tr), "FOREIGN KEY")
	tr.ProjectID = p.ProjectID

	tr.EnvID = unexisting
	assert.ErrorContains(t, f.gormdb.createTrigger(f.ctx, &tr), "FOREIGN KEY")
	tr.EnvID = env.EnvID

	tr.ConnectionID = unexisting
	assert.ErrorContains(t, f.gormdb.createTrigger(f.ctx, &tr), "FOREIGN KEY")
	tr.ConnectionID = conn.ConnectionID

	// test with existing assets
	f.createTriggersAndAssert(t, tr)
}

func TestDeleteTrigger(t *testing.T) {
	f := preTriggerTest(t, true) // no foreign key

	tr := f.newTrigger()
	f.createTriggersAndAssert(t, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	f.assertTriggersDeleted(t, tr)
}
