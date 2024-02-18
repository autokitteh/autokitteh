package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) SaveSignal(ctx context.Context, signalID string, workflowID string, connectionID sdktypes.ConnectionID, eventName string) (string, error) {
	s := scheme.Signal{
		ConnectionID: connectionID.String(),
		SignalID:     signalID,
		WorkflowID:   workflowID,
		EventType:    eventName,
	}

	return signalID, db.db.Create(&s).Error
}

func (db *gormdb) ListSignalsWaitingOnConnection(ctx context.Context, connectionID sdktypes.ConnectionID, eventType string) ([]scheme.Signal, error) {
	var signals []scheme.Signal
	q := db.db.WithContext(ctx).Where("connection_id = ?", connectionID.String()).Where("event_type = ?", eventType)
	if err := q.Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}

func (db *gormdb) RemoveSignal(ctx context.Context, signalID string) error {
	return db.db.WithContext(ctx).Where("signal_id = ?", signalID).Delete(&scheme.Signal{}).Error
}

func (db *gormdb) GetSignal(ctx context.Context, signalID string) (scheme.Signal, error) {
	var signal scheme.Signal
	if err := db.db.WithContext(ctx).Preload("Connection").Where("signal_id = ?", signalID).First(&signal).Error; err != nil {
		return signal, err
	}
	return signal, nil
}
