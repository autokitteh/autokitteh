package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) saveSignal(ctx context.Context, signal *scheme.Signal) error {
	return db.db.WithContext(ctx).Create(signal).Error
}

func (db *gormdb) SaveSignal(ctx context.Context, signalID string, workflowID string, connectionID sdktypes.ConnectionID, filter string) (string, error) {
	s := scheme.Signal{
		ConnectionID: *connectionID.Value(),
		SignalID:     signalID,
		WorkflowID:   workflowID,
		Filter:       filter,
	}

	return signalID, translateError(db.saveSignal(ctx, &s))
}

func (db *gormdb) ListSignalsWaitingOnConnection(ctx context.Context, connectionID sdktypes.ConnectionID) ([]scheme.Signal, error) {
	var signals []scheme.Signal
	q := db.db.WithContext(ctx).Where("connection_id = ?", connectionID.Value())
	if err := q.Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

func (db *gormdb) RemoveSignal(ctx context.Context, signalID string) error {
	return db.db.WithContext(ctx).Delete(&scheme.Signal{SignalID: signalID}).Error
}

func (db *gormdb) GetSignal(ctx context.Context, signalID string) (scheme.Signal, error) {
	var signal scheme.Signal
	if err := db.db.WithContext(ctx).Preload("Connection").Where("signal_id = ?", signalID).First(&signal).Error; err != nil {
		return signal, err
	}
	return signal, nil
}
