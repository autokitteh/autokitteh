package dbgorm

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

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
	return scheme.Owned{OwnerUserID: o.OwnerID().UUIDValue()}
}

func withOwnerID(q *gorm.DB, oid sdktypes.OwnerID) *gorm.DB {
	if oid.IsValid() {
		return q.Where("owner_user_id = ?", oid.UUIDValue())
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
	q = q.Where("projects.owner_user_id = ?", oid.UUIDValue())

	return q
}
