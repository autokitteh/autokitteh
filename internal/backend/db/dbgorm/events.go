package dbgorm

import (
	"context"
	"encoding/json"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) saveEvent(ctx context.Context, event *scheme.Event) error {
	return db.db.WithContext(ctx).Create(&event).Error
}

func (db *gormdb) SaveEvent(ctx context.Context, event sdktypes.Event) error {
	if err := event.Strict(); err != nil {
		return err
	}

	cid := event.ConnectionID()

	conn, err := db.GetConnection(ctx, cid)
	if err != nil {
		return err
	}

	if !conn.IsValid() {
		return sdkerrors.NewInvalidArgumentError("event connection_id does not exist")
	}

	e := scheme.Event{
		EventID:       event.ID().UUIDValue(),
		IntegrationID: conn.IntegrationID().UUIDValue(),
		ConnectionID:  cid.UUIDValue(),
		EventType:     event.Type(),
		Data:          kittehs.Must1(json.Marshal(event.Data())),
		Memo:          kittehs.Must1(json.Marshal(event.Memo())),
		CreatedAt:     event.CreatedAt(),
	}
	return translateError(db.saveEvent(ctx, &e))
}

func (db *gormdb) deleteEvent(ctx context.Context, id sdktypes.UUID) error {
	return db.db.WithContext(ctx).Delete(&scheme.Event{}, "event_id = ?", id).Error
}

func (db *gormdb) GetEventByID(ctx context.Context, eventID sdktypes.EventID) (sdktypes.Event, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseEvent, "event_id = ?", eventID.UUIDValue())
}

func (db *gormdb) ListEvents(ctx context.Context, filter sdkservices.ListEventsFilter) ([]sdktypes.Event, error) {
	q := db.db.WithContext(ctx)
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

	q = q.Where("seq >= ?", filter.MinSequenceNumber)

	q = q.Order("created_at asc") // hard coded now to get oldest first to support workflow events

	var es []scheme.Event
	if err := q.Find(&es).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(es, scheme.ParseEvent)
}

func (db *gormdb) GetLatestEventSequence(ctx context.Context) (uint64, error) {
	var s scheme.Event
	if err := db.db.WithContext(ctx).Last(&s).Error; err != nil {
		return 0, err
	}

	return s.Seq, nil
}
