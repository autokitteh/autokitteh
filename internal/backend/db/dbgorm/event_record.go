package dbgorm

import (
	"context"
	"time"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) addEventRecord(ctx context.Context, er *scheme.EventRecord) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(ctx, er.EventID); err != nil {
			return gormErrNotFoundToForeignKey(err) // should be present
		}
		var seq int64
		if err := tx.db.Model(&scheme.EventRecord{}).
			Where("event_id = ?", er.EventID).Count(&seq).Error; err != nil {
			return err
		}
		er.Seq = uint32(seq)
		er.CreatedAt = time.Now()
		return tx.db.Create(&er).Error
	})
}

func (gdb *gormdb) listEventRecords(ctx context.Context, eventID sdktypes.UUID) ([]scheme.EventRecord, error) {
	var ers []scheme.EventRecord
	if err := gdb.transaction(ctx, func(tx *tx) error { // REVIEW: do we need transaction in those cases?
		if err := tx.isCtxUserEntity(ctx, eventID); err != nil {
			return err
		}
		return tx.db.Model(&scheme.EventRecord{}).
			Where("event_id = ?", eventID).Order("event_id DESC, seq DESC").Find(&ers).Error
	}); err != nil {
		return nil, err
	}
	return ers, nil
}

// ------------------------------------------------------------------------------------------------
func (db *gormdb) AddEventRecord(ctx context.Context, er sdktypes.EventRecord) error {
	if err := er.Strict(); err != nil {
		return err
	}

	e := scheme.EventRecord{
		EventID: er.EventID().UUIDValue(),
		State:   int32(er.State().ToProto()),
	}

	return translateError(db.addEventRecord(ctx, &e))
}

func (db *gormdb) ListEventRecords(ctx context.Context, filter sdkservices.ListEventRecordsFilter) ([]sdktypes.EventRecord, error) {
	// NOTE: invalid eventID won't list all events for all users
	eventRecords, err := db.listEventRecords(ctx, filter.EventID.UUIDValue())
	if eventRecords == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(eventRecords, scheme.ParseEventRecord)
}
