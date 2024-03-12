package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func createTriggersAndAssert(t *testing.T, f *dbFixture, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assert.NoError(t, f.gormdb.createTrigger(f.ctx, trigger))
		findAndAssertOne(t, f, trigger, "trigger_id = ?", trigger.TriggerID)
	}
}

func assertTriggersDeleted(t *testing.T, f *dbFixture, triggers ...scheme.Trigger) {
	for _, trigger := range triggers {
		assertDeleted(t, f, scheme.Trigger{TriggerID: trigger.TriggerID})
	}
}

func TestCreateTrigger(t *testing.T) {
	f := newDbFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Trigger{}, 0, "") // no trigger

	tr := newTrigger(f)
	// test createTrigger
	createTriggersAndAssert(t, f, tr)
}

func TestDeleteTrigger(t *testing.T) {
	f := newDbFixture(true)                           // no foreign keys
	findAndAssertCount(t, f, scheme.Trigger{}, 0, "") // no trigger

	tr := newTrigger(f)
	createTriggersAndAssert(t, f, tr)

	// test deleteTrigger
	assert.NoError(t, f.gormdb.deleteTrigger(f.ctx, tr.TriggerID))
	assertTriggersDeleted(t, f, tr)
}
