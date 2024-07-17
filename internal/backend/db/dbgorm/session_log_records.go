package dbgorm

import (
	"context"
	"encoding/json"
	"fmt"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) SaveSessionLogRecord(ctx context.Context, sessionID sdktypes.SessionID, record sdktypes.SessionLogRecord) error {
	record = record.WithProcessID(fixtures.ProcessID())
	logRecordData, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal session log record: %w", err)
	}

	fmt.Println("Seq", record.Seq())
	return db.db.Create(&scheme.SessionLogRecord{SessionID: sessionID.UUIDValue(), Data: logRecordData, Seq: uint64(record.Seq())}).Error
}
