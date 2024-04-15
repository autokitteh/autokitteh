package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (f *dbFixture) createEventRecordsAndAssert(t *testing.T, eventRecords ...scheme.EventRecord) {
	for _, er := range eventRecords {
		assert.NoError(t, f.gormdb.addEventRecord(f.ctx, &er))
		findAndAssertOne(t, f, er, "event_id = ?", er.EventID)
	}
}

func TestCreateEventRecordForeignKeys(t *testing.T) {
	f := newDBFixture()
	findAndAssertCount(t, f, scheme.EventRecord{}, 0, "") // no events

	evt := f.newEvent()
	er := f.newEventRecord()
	assert.ErrorIs(t, f.gormdb.addEventRecord(f.ctx, &er), gorm.ErrForeignKeyViolated)

	f.createEventsAndAssert(t, evt)
	// test createEventRecord
	f.createEventRecordsAndAssert(t, er)
}
