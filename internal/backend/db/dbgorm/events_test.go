package dbgorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
	f := newDBFixture(false)
	findAndAssertCount(t, f, scheme.Event{}, 0, "") // no events

	evt := f.newEvent()
	// test createEvent
	f.createEventsAndAssert(t, evt)
}

func TestCreateEventForeignKeys(t *testing.T) {
	f := newDBFixture(false)
	findAndAssertCount(t, f, scheme.Event{}, 0, "") // no events

	e := f.newEvent()
	unexisting := "unexisting"

	e.IntegrationID = &unexisting
	assert.ErrorContains(t, f.gormdb.saveEvent(f.ctx, &e), "FOREIGN KEY")
	e.IntegrationID = nil

	// FIXME: ENG-569
	// e.OriginalEventID = &unexisting
	// assert.ErrorContains(t, f.gormdb.saveEvent(f.ctx, &e), "FOREIGN KEY")
	// e.OriginalEventID = nil

	// e2 := f.newEvent()
	// f.createEventsAndAssert(t, e2)
	// e.OriginalEventID = &e2.EventID

	i := f.newIntegration()
	f.createIntegrationsAndAssert(t, i)

	e.IntegrationID = &i.IntegrationID
	f.createEventsAndAssert(t, e)
}

func TestDeleteEvent(t *testing.T) {
	f := newDBFixture(true)                         // no foreign keys
	findAndAssertCount(t, f, scheme.Event{}, 0, "") // no events

	evt := f.newEvent()
	f.createEventsAndAssert(t, evt)

	// test deleteEvent
	assert.NoError(t, f.gormdb.deleteEvent(f.ctx, evt.EventID))
	f.assertEventsDeleted(t, evt)
}
