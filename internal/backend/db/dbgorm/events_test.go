package dbgorm

import (
	"context"
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
	findAndAssertCount[scheme.Event](t, f, 0, "") // no events

	evt := f.newEvent() // connection is nil
	f.createEventsAndAssert(t, evt)

	// test deleteEvent
	assert.NoError(t, f.gormdb.deleteEvent(f.ctx, evt.EventID))
	f.assertEventsDeleted(t, evt)
}

func TestListEventsDefaultOrder(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false) // no foreign keys

	ctx := context.Background()
	ev1 := f.newEvent()
	ev2 := f.newEvent()
	f.createEventsAndAssert(t, ev1, ev2)

	evs, err := f.gormdb.ListEvents(ctx, sdkservices.ListEventsFilter{})

	require.NoError(t, err)

	require.Equal(t, 2, len(evs), "should be 2 events in db")

	// Desc order is default
	require.Equal(t, evs[0].ID().UUIDValue(), ev2.EventID)
	require.Equal(t, evs[1].ID().UUIDValue(), ev1.EventID)
}

func TestListEventsOrder(t *testing.T) {
	f := newDBFixture()
	foreignKeys(f.gormdb, false) // no foreign keys

	ctx := context.Background()
	ev1 := f.newEvent()
	ev2 := f.newEvent()
	f.createEventsAndAssert(t, ev1, ev2)

	tests := []struct {
		name  string
		order sdkservices.ListOrder
		ids   [2]uuid.UUID
	}{
		{
			name:  "descending order",
			order: sdkservices.ListOrderDescending,
			ids:   [2]uuid.UUID{ev2.EventID, ev1.EventID},
		},
		{
			name:  "ascending order",
			order: sdkservices.ListOrderAscending,
			ids:   [2]uuid.UUID{ev1.EventID, ev2.EventID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evs, err := f.gormdb.ListEvents(ctx, sdkservices.ListEventsFilter{Order: tt.order})
			require.NoError(t, err)
			require.Equal(t, 2, len(evs), "should be 2 events in db")
			require.Equal(t, tt.ids, [2]uuid.UUID{evs[0].ID().UUIDValue(), evs[1].ID().UUIDValue()})
		})
	}
}
