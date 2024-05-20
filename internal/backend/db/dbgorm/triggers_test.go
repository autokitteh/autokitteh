package dbgorm

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func createTrigger(ctx context.Context, db *gormdb, trigger *scheme.Trigger) error {
	if trigger.Name == "" {
		// This protects against triggers with empty names, which will cause unique validation.
		// Happens mostly in testing.
		trigger.UniqueName = uuid.New().String()
	}

	return db.createTrigger(ctx, trigger)
}

func (f *dbFixture) createTriggersAndAssert(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assert.NoError(t, createTrigger(f.ctx, f.gormdb, &trigger))
		findAndAssertOne(t, f, trigger, "trigger_id = ?", trigger.TriggerID)
	}
}

func (f *dbFixture) assertTriggersDeleted(t *testing.T, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assertDeleted(t, f, trigger)
	}
}

func TestCreateTrigger(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false)                    // no foreign keys
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers

	tr := f.newTrigger()
	// test createTrigger
	f.createTriggersAndAssert(t, tr)
}

func TestCreateTriggerForeignKeys(t *testing.T) {
	f := newDBFixture()
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers

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
	unexisting := uuid.New()

	tr.ProjectID = unexisting
	assert.ErrorIs(t, createTrigger(f.ctx, f.gormdb, &tr), gorm.ErrForeignKeyViolated)
	tr.ProjectID = p.ProjectID

	tr.EnvID = unexisting
	assert.ErrorIs(t, createTrigger(f.ctx, f.gormdb, &tr), gorm.ErrForeignKeyViolated)
	tr.EnvID = env.EnvID

	tr.ConnectionID = unexisting
	assert.ErrorIs(t, createTrigger(f.ctx, f.gormdb, &tr), gorm.ErrForeignKeyViolated)
	tr.ConnectionID = conn.ConnectionID

	// test with existing assets
	f.createTriggersAndAssert(t, tr)
}

func TestDeleteTrigger(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false)                    // no foreign keys
	findAndAssertCount[scheme.Trigger](t, f, 0, "") // no triggers

	tr := f.newTrigger()
	f.createTriggersAndAssert(t, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	f.assertTriggersDeleted(t, tr)
}
