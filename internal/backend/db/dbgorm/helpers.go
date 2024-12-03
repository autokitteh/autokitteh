package dbgorm

import (
	"context"
	"fmt"
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

func ownedBy(o interface{ OwnerID() sdktypes.OwnerID }) scheme.Owned {
	uuid := o.OwnerID().UUIDValue()

	r := scheme.Owned{OwnerID: uuid}

	if o.OwnerID().IsUserID() {
		r.OwnerUserID = uuid
	} else if o.OwnerID().IsOrgID() {
		r.OwnerOrgID = uuid
	}

	return r
}

func withOwnerID(q *gorm.DB, oid sdktypes.OwnerID) *gorm.DB {
	if oid.IsValid() {
		return q.Where("owner_id = ?", oid.UUIDValue())
	}

	return q
}

// ---

func belongsToProjectID(p sdktypes.ProjectID) scheme.BelongsToProject {
	return scheme.BelongsToProject{ProjectID: p.UUIDValue()}
}

func belongsToProjectIDOf(o interface{ ProjectID() sdktypes.ProjectID }) scheme.BelongsToProject {
	return belongsToProjectID(o.ProjectID())
}

func withProjectOwnerID(q *gorm.DB, targetTableName string, oid sdktypes.OwnerID) *gorm.DB {
	if !oid.IsValid() {
		return q
	}

	q = q.Preload("Project")
	q = q.Joins(fmt.Sprintf("JOIN projects ON projects.project_id = %s.project_id", targetTableName))
	q = q.Where("projects.owner_id = ?", oid.UUIDValue())

	return q
}

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
