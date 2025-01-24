package dbgorm

import (
	"context"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/types"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) saveSignal(ctx context.Context, signal *scheme.Signal) error {
	return db.wdb.WithContext(ctx).Create(signal).Error
}

func (db *gormdb) SaveSignal(ctx context.Context, signal *types.Signal) error {
	s := scheme.Signal{
		DestinationID: signal.DestinationID.UUIDValue(),
		ConnectionID:  signal.DestinationID.ToConnectionID().UUIDValuePtr(),
		TriggerID:     signal.DestinationID.ToTriggerID().UUIDValuePtr(),
		SignalID:      signal.ID,
		WorkflowID:    signal.WorkflowID,
		Filter:        signal.Filter,
	}

	return translateError(db.saveSignal(ctx, &s))
}

func (db *gormdb) ListWaitingSignals(ctx context.Context, dstID sdktypes.EventDestinationID) ([]*types.Signal, error) {
	var rs []*scheme.Signal
	q := db.rdb.WithContext(ctx).Where("destination_id = ?", dstID.UUIDValue())
	if err := q.Find(&rs).Error; err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(rs, scheme.ParseSignal)
}

func (db *gormdb) RemoveSignal(ctx context.Context, signalID uuid.UUID) error {
	return translateError(db.wdb.WithContext(ctx).Delete(&scheme.Signal{SignalID: signalID}).Error)
}

func (db *gormdb) GetSignal(ctx context.Context, signalID uuid.UUID) (*types.Signal, error) {
	var signal scheme.Signal

	q := db.rdb.
		WithContext(ctx).
		Where("signal_id = ?", signalID).
		Preload("Connection").
		Preload("Trigger")

	if err := q.First(&signal).Error; err != nil {
		return nil, translateError(err)
	}

	return scheme.ParseSignal(&signal)
}
