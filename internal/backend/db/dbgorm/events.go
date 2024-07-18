package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) withUserEvents(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "event")
}

func (gdb *gormdb) saveEvent(ctx context.Context, event *scheme.Event) error {
	cronConnectionID := sdktypes.BuiltinSchedulerConnectionID.UUIDValue()
	if event.ConnectionID == nil || event.ConnectionID == &cronConnectionID {
		return fmt.Errorf("find event owner: bad connection %s", event.ConnectionID)
	}
	if !authcontext.NeedOwnershipCheck(ctx) {
		ctx = gdb.CtxWithOwnershipOf(ctx, *event.ConnectionID)
	}
	createFunc := func(tx *gorm.DB, uid string) error {
		if event.ConnectionID != nil && *event.ConnectionID != cronConnectionID {
			var con scheme.Connection
			if err := tx.Where("connection_id = ?", event.ConnectionID).First(&con).Error; err != nil {
				return fmt.Errorf("find event connection: %w", err)
			}
			if con.IntegrationID != nil {
				event.IntegrationID = con.IntegrationID
			}
		}
		return tx.Create(event).Error
	}
	return gdb.createEntityWithOwnership(ctx, createFunc, event, event.ConnectionID)
}

func (gdb *gormdb) deleteEvent(ctx context.Context, eventID sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(ctx, eventID); err != nil {
			return err
		}
		return tx.db.Delete(&scheme.Event{}, "event_id = ?", eventID).Error // NOTE: eventID isn't a primary key
	})
}

func (gdb *gormdb) getEvent(ctx context.Context, eventID sdktypes.UUID) (*scheme.Event, error) {
	return getOne[scheme.Event](gdb.withUserEvents(ctx), "event_id = ?", eventID)
}

func (gdb *gormdb) listEvents(ctx context.Context, filter sdkservices.ListEventsFilter) ([]scheme.Event, error) {
	q := gdb.withUserEvents(ctx)

	if filter.IntegrationID.IsValid() {
		q = q.Where("integration_id = ?", filter.IntegrationID.UUIDValue())
	}
	if filter.ConnectionID.IsValid() {
		q = q.Where("connection_id = ?", filter.ConnectionID.UUIDValue())
	}
	if filter.EventType != "" {
		q = q.Where("event_type = ?", filter.EventType)
	}
	if filter.CreatedAfter != nil {
		q = q.Where("created_at > ?", filter.CreatedAfter)
	}

	if filter.Limit != 0 {
		q = q.Limit(filter.Limit)
	}

	q = q.Where("seq > ?", filter.MinSequenceNumber)

	if filter.Order == sdkservices.ListOrderAscending {
		q = q.Order("seq asc")
	} else {
		q = q.Order("seq desc") // default to desc
	}

	var es []scheme.Event
	if err := q.Omit("data").Find(&es).Error; err != nil {
		return nil, err
	}
	return es, nil
}

func (db *gormdb) sdkToRecord(event sdktypes.Event) (*scheme.Event, error) {
	if err := event.Strict(); err != nil {
		return nil, err
	}

	e := scheme.Event{
		EventID:      event.ID().UUIDValue(),
		ConnectionID: scheme.UUIDOrNil(event.ConnectionID().UUIDValue()), // could be invalid/zero
		EventType:    event.Type(),
		Data:         kittehs.Must1(json.Marshal(event.Data())),
		Memo:         kittehs.Must1(json.Marshal(event.Memo())),
		CreatedAt:    event.CreatedAt(),
	}
	return &e, nil
}

func (db *gormdb) SaveEvent(ctx context.Context, event sdktypes.Event) error {
	e, err := db.sdkToRecord(event)
	if err != nil {
		return err
	}
	return translateError(db.saveEvent(ctx, e))
}

func (db *gormdb) GetEventByID(ctx context.Context, eventID sdktypes.EventID) (sdktypes.Event, error) {
	e, err := db.getEvent(ctx, eventID.UUIDValue())
	if e == nil || err != nil {
		return sdktypes.InvalidEvent, translateError(err)
	}
	return scheme.ParseEvent(*e)
}

func (db *gormdb) ListEvents(ctx context.Context, filter sdkservices.ListEventsFilter) ([]sdktypes.Event, error) {
	events, err := db.listEvents(ctx, filter)
	if events == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(events, scheme.ParseEvent)
}

func (db *gormdb) GetLatestEventSequence(ctx context.Context) (uint64, error) {
	// NOTE: called from workflow, not protected by user context
	var s scheme.Event
	if err := db.db.WithContext(ctx).Last(&s).Error; err != nil {
		return 0, err
	}
	return s.Seq, nil
}
