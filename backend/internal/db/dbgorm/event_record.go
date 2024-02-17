package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) AddEventRecord(ctx context.Context, er sdktypes.EventRecord) error {
	e := scheme.EventRecord{
		Seq:     sdktypes.GetEventRecordSeq(er),
		EventID: sdktypes.GetEventRecordEventID(er).String(),
		State:   int32(sdktypes.GetEventRecordState(er)),
	}

	if err := db.db.WithContext(ctx).Create(&e).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) ListEventRecords(ctx context.Context, filter sdkservices.ListEventRecordsFilter) ([]sdktypes.EventRecord, error) {
	q := db.db.WithContext(ctx)
	if filter.EventID != nil {
		q = q.Where("event_id = ?", filter.EventID.String())
	}

	q.Order("event_id DESC, seq DESC")

	var ers []scheme.EventRecord
	if err := q.Find(&ers).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(ers, scheme.ParseEventRecord)
}
