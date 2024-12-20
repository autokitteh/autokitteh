package dbgorm

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func uuidPtrOrNil(o interface{ UUIDValue() uuid.UUID }) *uuid.UUID {
	v := o.UUIDValue()

	if v == uuid.Nil {
		return nil
	}

	return &v
}

func withProjectID(q *gorm.DB, field string, pid sdktypes.ProjectID) *gorm.DB {
	if pid.IsValid() {
		if field != "" {
			field += "."
		}

		return q.Where(field+"project_id = ?", pid.UUIDValue())
	}

	return q
}

// ---

func based(ctx context.Context) scheme.Base {
	now := time.Now().UTC()

	uid := authcontext.GetAuthnUserID(ctx).UUIDValue()

	return scheme.Base{
		CreatedBy: uid,
		CreatedAt: now,
	}
}

func updatedBaseColumns(ctx context.Context) map[string]any {
	return map[string]any{
		"updated_at": time.Now().UTC(),
		"updated_by": authcontext.GetAuthnUserID(ctx).UUIDValue(),
	}
}
