package dbgorm

import (
	"context"
	"encoding/json"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) SaveEvent(ctx context.Context, event sdktypes.Event) error {
	e := scheme.Event{
		EventID:          event.ID().String(),
		IntegrationID:    event.IntegrationID().String(), // TODO(ENG-158): need to verify integration id
		IntegrationToken: event.IntegrationToken(),
		OriginalEventID:  event.OriginalEventID(),
		EventType:        event.Type(),
		Data:             kittehs.Must1(json.Marshal(event.Data())),
		Memo:             kittehs.Must1(json.Marshal(event.Memo())),
		CreatedAt:        event.CreatedAt(),
	}

	if err := db.db.WithContext(ctx).Create(&e).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) GetEventByID(ctx context.Context, eventID sdktypes.EventID) (sdktypes.Event, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseEvent, "event_id = ?", eventID.String())
}

func (db *gormdb) ListEvents(ctx context.Context, filter sdkservices.ListEventsFilter) ([]sdktypes.Event, error) {
	q := db.db.WithContext(ctx)
	if filter.IntegrationID.IsValid() {
		q = q.Where("integration_id = ?", filter.IntegrationID.String())
	}

	if filter.IntegrationToken != "" {
		q = q.Where("integration_token = ?", filter.IntegrationToken)
	}

	if filter.EventType != "" {
		q = q.Where("event_type = ?", filter.EventType)
	}

	if filter.OriginalID != "" {
		q = q.Where("original_id = ?", filter.OriginalID)
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
