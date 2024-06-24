package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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

func preTriggerTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers

	p := f.newProject() // parent project
	f.createProjectsAndAssert(t, p)
	f.projectID = p.ProjectID
	return f
}

func TestCreateTrigger(t *testing.T) {
	f := preTriggerTest(t)
	foreignKeys(f.gormdb, false) // no foreign keys

	tr := f.newTrigger()

	// NOTE: this test won't fail due to foreign key violation of trigger.projectID
	// But on user scope check (for unexisting projectID, which is not in DB and thus not in user scope)
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr), sdkerrors.ErrUnauthorized)

	// test createTrigger
	tr.ProjectID = f.projectID
	f.createTriggersAndAssert(t, tr)
}

func TestGetTrigger(t *testing.T) {
	f := preTriggerTest(t).WithDebug()
	foreignKeys(f.gormdb, false) // no foreign keys

	tr := f.newTrigger()
	tr.ProjectID = f.projectID
	f.createTriggersAndAssert(t, tr)

	// test getTrigger
	tr2, err := f.gormdb.getTrigger(f.ctx, tr.TriggerID)
	assert.NoError(t, err)
	assert.Equal(t, tr, *tr2)
}

func TestCreateTriggerForeignKeys(t *testing.T) {
	f := preTriggerTest(t)
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers

	// prepare
	tr := f.newTrigger()
	// p := f.newProject(). Already created in preTriggerTest
	env := f.newEnv()
	conn := f.newConnection()

	env.ProjectID = f.projectID

	tr.ProjectID = f.projectID
	tr.EnvID = env.EnvID
	tr.ConnectionID = conn.ConnectionID

	// f.createProjectsAndAssert(t, p)
	f.createEnvsAndAssert(t, env)
	f.createConnectionsAndAssert(t, conn)

	// negative test with non-existing assets
	unexisting := uuid.New()

	tr.EnvID = unexisting
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr), gorm.ErrForeignKeyViolated)
	tr.EnvID = env.EnvID

	tr.ConnectionID = unexisting
	assert.ErrorIs(t, f.gormdb.createTrigger(f.ctx, &tr), gorm.ErrForeignKeyViolated)
	tr.ConnectionID = conn.ConnectionID

	// test with existing assets
	f.createTriggersAndAssert(t, tr)
}

func TestDeleteTrigger(t *testing.T) {
	f := preTriggerTest(t)
	foreignKeys(f.gormdb, false) // no foreign keys

	tr := f.newTrigger()
	tr.ProjectID = f.projectID
	f.createTriggersAndAssert(t, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	f.assertTriggersDeleted(t, tr)
}
