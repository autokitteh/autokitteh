package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (f *dbFixture) createEventsAndAssert(t *testing.T, events ...scheme.Event) {
	for _, event := range events {
		assert.NoError(t, f.gormdb.saveEvent(f.ctx, &event))
		findAndAssertOne(t, f, event, "event_id = ?", event.EventID)
	}
}

func (f *dbFixture) assertEventsDeleted(t *testing.T, events ...scheme.Event) {
	for _, event := range events {
		assertSoftDeleted(t, f, scheme.Event{EventID: event.EventID})
	}
}

func TestCreateEvent(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false)                  // no foreign keys
	findAndAssertCount[scheme.Event](t, f, 0, "") // no events

	evt := f.newEvent()
	// test createEvent
	f.createEventsAndAssert(t, evt)
}

func TestCreateEventForeignKeys(t *testing.T) {
	f := newDBFixture()
	findAndAssertCount[scheme.Event](t, f, 0, "") // no events

	e := f.newEvent()
	i := f.newIntegration()
	c := f.newConnection()

	f.createIntegrationsAndAssert(t, i)
	f.createConnectionsAndAssert(t, c)

	// negative test with non-existing assets
	unexisting := uuid.New()

	// FIXME: ENG-590. foreign keys integration
	// e.IntegrationID = &unexisting
	// assert.ErrorIs(t, f.gormdb.saveEvent(f.ctx, &e), gorm.ErrForeignKeyViolated)
	// e.IntegrationID = &i.IntegrationID

	e.ConnectionID = &unexisting
	assert.ErrorIs(t, f.gormdb.saveEvent(f.ctx, &e), gorm.ErrForeignKeyViolated)
	e.ConnectionID = &c.ConnectionID

	f.createEventsAndAssert(t, e)
}

func TestDeleteEvent(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false)                  // no foreign keys
	findAndAssertCount[scheme.Event](t, f, 0, "") // no events

	evt := f.newEvent()
	f.createEventsAndAssert(t, evt)

	// test deleteEvent
	assert.NoError(t, f.gormdb.deleteEvent(f.ctx, evt.EventID))
	f.assertEventsDeleted(t, evt)
}
