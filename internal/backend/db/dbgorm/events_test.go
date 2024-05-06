package dbgorm

import (
	"testing"

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
	f := newDBFixtureFK(true)                       // no foreign keys
	findAndAssertCount(t, f, scheme.Event{}, 0, "") // no events

	evt := f.newEvent()
	// test createEvent
	f.createEventsAndAssert(t, evt)
}

func TestCreateEventForeignKeys(t *testing.T) {
	f := newDBFixtureFK(false)
	findAndAssertCount[scheme.Event](t, f, 0, "") // no events

	e := f.newEvent()
	assert.ErrorIs(t, f.gormdb.saveEvent(f.ctx, &e), gorm.ErrForeignKeyViolated)

	i := f.newIntegration()
	f.createIntegrationsAndAssert(t, i)
	e.IntegrationID = i.IntegrationID
	assert.ErrorIs(t, f.gormdb.saveEvent(f.ctx, &e), gorm.ErrForeignKeyViolated) // need connection as well

	c := f.newConnection()
	f.createConnectionsAndAssert(t, c)
	e.ConnectionID = c.ConnectionID
	f.createEventsAndAssert(t, e)
}

func TestDeleteEvent(t *testing.T) {
	f := newDBFixtureFK(true)                     // no foreign keys
	findAndAssertCount[scheme.Event](t, f, 0, "") // no events

	evt := f.newEvent()
	f.createEventsAndAssert(t, evt)

	// test deleteEvent
	assert.NoError(t, f.gormdb.deleteEvent(f.ctx, evt.EventID))
	f.assertEventsDeleted(t, evt)
}
