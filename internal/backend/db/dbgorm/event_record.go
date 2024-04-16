package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) addEventRecord(ctx context.Context, er *scheme.EventRecord) error {
	return db.db.WithContext(ctx).Create(&er).Error
}

func (db *gormdb) AddEventRecord(ctx context.Context, er sdktypes.EventRecord) error {
	e := scheme.EventRecord{
		Seq:     er.Seq(),
		EventID: *er.EventID().Value(),
		State:   int32(er.State().ToProto()),
	}

	return translateError(db.addEventRecord(ctx, &e))
}

func (db *gormdb) ListEventRecords(ctx context.Context, filter sdkservices.ListEventRecordsFilter) ([]sdktypes.EventRecord, error) {
	q := db.db.WithContext(ctx)
	if filter.EventID.IsValid() {
		q = q.Where("event_id = ?", filter.EventID.Value())
	}

	q.Order("event_id DESC, seq DESC")

	var ers []scheme.EventRecord
	if err := q.Find(&ers).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(ers, scheme.ParseEventRecord)
}
