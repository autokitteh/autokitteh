package dbgorm

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
)

func (db *gormdb) addUserAuditLog(ctx context.Context, description string, data any, err error) {
	uidStr := authcontext.GetAuthnUserID(ctx)

	l := db.z.With(zap.String("uid", uidStr), zap.String("description", description), zap.Error(err))

	jsonData, err := json.Marshal(data)
	if err != nil {
		l.DPanic("failed to marshal data", zap.Error(err))
		jsonData = []byte("null")
	}

	var errStr string
	if err != nil {
		errStr = err.Error()
	}

	uuid, err := uuid.Parse(uidStr)
	if err != nil {
		l.DPanic("failed to parse uuid", zap.Error(err))
		return
	}

	r := scheme.UserAuditLog{
		UserID:      uuid,
		Description: description,
		Data:        jsonData,
		Timestamp:   time.Now().UTC(),
		Error:       errStr,
	}

	if err := db.db.WithContext(ctx).Create(&r).Error; err != nil {
		l.Error("failed to create user audit log", zap.Error(err))
	}
}
