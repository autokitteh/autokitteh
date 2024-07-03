package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

// func (f *dbFixture) createEventRecordsAndAssert(t *testing.T, eventRecord scheme.EventRecord) {
// 	assert.NoError(t, f.gormdb.addEventRecord(f.ctx, &eventRecord))
// 	findAndAssertOne(t, f, er, "event_id = ?", eventRecord.EventID)
// }

func preEventRecordTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.EventRecord](t, f, 0, "") // no event records

	e := f.newEvent()
	f.createEventsAndAssert(t, e)
	f.eventID = e.EventID
	return f
}

func TestEventRecord(t *testing.T) {
	f := preEventRecordTest(t)

	er := f.newEventRecord()
	assert.ErrorIs(t, f.gormdb.addEventRecord(f.ctx, &er), gorm.ErrForeignKeyViolated)

	// test createEventRecord
	er.EventID = f.eventID

	// test listEventRecords. Add 2 event records, check if they are listed and their seq
	for i := 0; i < 2; i++ {
		assert.NoError(t, f.gormdb.addEventRecord(f.ctx, &er))

		ers, err := f.gormdb.listEventRecords(f.ctx, er.EventID)
		assert.NoError(t, err)

		assert.Equal(t, i+1, len(ers))              // how many even records
		assert.Equal(t, er.EventID, ers[0].EventID) // eventID of the last added record
		assert.Equal(t, uint32(i), ers[0].Seq)      // seq of the last added record
	}
}
