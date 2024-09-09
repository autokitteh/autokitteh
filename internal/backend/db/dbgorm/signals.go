package dbgorm

import (
	"context"

	"github.com/google/uuid"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) saveSignal(ctx context.Context, signal *scheme.Signal) error {
	return db.db.WithContext(ctx).Create(signal).Error
}

func (db *gormdb) SaveSignal(ctx context.Context, signalID uuid.UUID, workflowID string, dstID sdktypes.EventDestinationID, filter string) error {
	s := scheme.Signal{
		DestinationID: dstID.UUIDValue(),
		ConnectionID:  dstID.ToConnectionID().UUIDValuePtr(),
		TriggerID:     dstID.ToTriggerID().UUIDValuePtr(),
		SignalID:      signalID,
		WorkflowID:    workflowID,
		Filter:        filter,
	}

	return translateError(db.saveSignal(ctx, &s))
}

func (db *gormdb) ListWaitingSignals(ctx context.Context, dstID sdktypes.EventDestinationID) ([]scheme.Signal, error) {
	var signals []scheme.Signal
	q := db.db.WithContext(ctx).Where("destination_id = ?", dstID.UUIDValue())
	if err := q.Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

func (db *gormdb) RemoveSignal(ctx context.Context, signalID uuid.UUID) error {
	return db.db.WithContext(ctx).Delete(&scheme.Signal{SignalID: signalID}).Error
}

func (db *gormdb) GetSignal(ctx context.Context, signalID uuid.UUID) (scheme.Signal, error) {
	var signal scheme.Signal

	q := db.db.
		WithContext(ctx).
		Where("signal_id = ?", signalID).
		Preload("Connection").
		Preload("Trigger")

	if err := q.First(&signal).Error; err != nil {
		return signal, err
	}
	return signal, nil
}
