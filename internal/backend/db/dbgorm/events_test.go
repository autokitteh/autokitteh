package dbgorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

func (f *dbFixture) createEventsAndAssert(t *testing.T, events ...scheme.Event) {
	for _, event := range events {
		assert.NoError(t, f.gormdb.saveEvent(f.ctx, &event))
		findAndAssertOne(t, f, event, "event_id = ?", event.EventID)
	}
}

func (f *dbFixture) assertEventsDeleted(t *testing.T, events ...scheme.Event) {
	for _, event := range events {
		assertDeleted(t, f, scheme.Event{EventID: event.EventID})
	}
}

func preEventTest(t *testing.T) *dbFixture {
	f := newDBFixture()
	findAndAssertCount[scheme.Event](t, f, 0, "") // no events
	return f
}

func TestCreateEvent(t *testing.T) {
	f := preEventTest(t)

	p := f.newProject()
	f.createProjectsAndAssert(t, p)
	e := f.newEvent(p)

	// test createEvent
	f.createEventsAndAssert(t, e)
}

func TestCreateEventForeignKeys(t *testing.T) {
	f := preEventTest(t)

	p := f.newProject()
	e := f.newEvent(p)
	c := f.newConnection(p)
	b := f.newBuild(p)

	f.createProjectsAndAssert(t, p)
	f.saveBuildsAndAssert(t, b)
	f.createConnectionsAndAssert(t, c)

	// negative test with non-existing assets
	// use unexisingID = buildID owned by user to pass user ownership test

	e.IntegrationID = &testIntegrationID

	e.ConnectionID = &b.BuildID // no such connectionID, since it's a buildID
	assert.ErrorIs(t, f.gormdb.saveEvent(f.ctx, &e), gorm.ErrForeignKeyViolated)
	e.ConnectionID = &c.ConnectionID

	f.createEventsAndAssert(t, e)
}

func TestDeleteEvent(t *testing.T) {
	f := preEventTest(t)
	p := f.newProject()
	f.createProjectsAndAssert(t, p)
	e := f.newEvent(p) // connection is nil
	f.createEventsAndAssert(t, e)

	// test deleteEvent
	assert.NoError(t, f.gormdb.deleteEvent(f.ctx, e.EventID))
	f.assertEventsDeleted(t, e)
}

func TestDeleteEventNotDeletingSessions(t *testing.T) {
	f := preEventTest(t)
	p := f.newProject()
	f.createProjectsAndAssert(t, p)
	e := f.newEvent(p) // connection is nil
	f.createEventsAndAssert(t, e)
	b := f.newBuild(p)
	f.saveBuildsAndAssert(t, b)
	s := f.newSession(e, p, b)
	f.createSessionsAndAssert(t, s)

	// test deleteEvent
	assert.NoError(t, f.gormdb.deleteEvent(f.ctx, e.EventID))
	f.assertEventsDeleted(t, e)

	ses, err := f.gormdb.getSession(f.ctx, s.SessionID)
	assert.NoError(t, err)
	assert.Equal(t, ses.SessionID, s.SessionID)
}

func TestGetEvent(t *testing.T) {
	f := preEventTest(t)
	p := f.newProject()
	f.createProjectsAndAssert(t, p)
	e := f.newEvent(p)
	f.createEventsAndAssert(t, e)

	// test getEvent
	e2, err := f.gormdb.getEvent(f.ctx, e.EventID)
	assert.NoError(t, err)

	assert.Equal(t, e, *e2)

	assert.NoError(t, f.gormdb.deleteEvent(f.ctx, e.EventID))
	_, err = f.gormdb.getTriggerByID(f.ctx, e.EventID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestListEventsOrder(t *testing.T) {
	f := preEventTest(t)
	p := f.newProject()
	f.createProjectsAndAssert(t, p)

	ev1 := f.newEvent(p)
	ev2 := f.newEvent(p)
	f.createEventsAndAssert(t, ev1, ev2)

	tests := []struct {
		name   string
		filter sdkservices.ListEventsFilter
		ids    [2]uuid.UUID
	}{
		{
			name:   "default order",
			filter: sdkservices.ListEventsFilter{},
			ids:    [2]uuid.UUID{ev2.EventID, ev1.EventID},
		},
		{
			name:   "descending order",
			filter: sdkservices.ListEventsFilter{Order: sdkservices.ListOrderDescending},
			ids:    [2]uuid.UUID{ev2.EventID, ev1.EventID},
		},
		{
			name:   "ascending order",
			filter: sdkservices.ListEventsFilter{Order: sdkservices.ListOrderAscending},
			ids:    [2]uuid.UUID{ev1.EventID, ev2.EventID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evs, err := f.gormdb.listEvents(f.ctx, tt.filter)
			require.NoError(t, err)
			require.Equal(t, 2, len(evs), "should be 2 events in db")
			require.Equal(t, tt.ids, [2]uuid.UUID{evs[0].EventID, evs[1].EventID})
		})
	}
}
