package dbgorm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) withUserConnections(ctx context.Context) *gorm.DB {
	return gdb.withUserEntity(ctx, "connection")
}

func (gdb *gormdb) createConnection(ctx context.Context, conn *scheme.Connection) error {
	createFunc := func(tx *gorm.DB, uid string) error {
		// ensure there is no connection with the same name for the same project
		var count int64
		if err := tx.
			Model(&scheme.Connection{}).
			Scopes(withUserEntity(ctx, gdb, "connection", uid)).
			Where("name = ?", conn.Name).Where("project_id = ? or project_id is NULL", conn.ProjectID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return gorm.ErrDuplicatedKey // active/non-deleted connection was found.
		}
		return tx.Create(conn).Error
	}
	return gdb.createEntityWithOwnership(ctx, createFunc, conn, conn.ProjectID)
}

func (gdb *gormdb) deleteConnectionsAndVars(what string, id sdktypes.UUID) error {
	// should be transactional with context already applied

	var ids []sdktypes.UUID
	q := gdb.db.Model(&scheme.Connection{})
	q = q.Clauses(clause.Returning{Columns: []clause.Column{{Name: "connection_id"}}})
	if err := q.Delete(&ids, fmt.Sprintf("%s = ?", what), id).Error; err != nil {
		return err
		// REVIEW: proceed to vars deletion if there are any?
	}

	if len(ids) > 0 {
		return gdb.db.Where("var_id IN (?)", ids).Delete(&scheme.Var{}).Error
	}
	return nil
}

func (gdb *gormdb) deleteConnection(ctx context.Context, id sdktypes.UUID) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(ctx, id); err != nil {
			return err
		}
		// delete connection and associated vars
		return tx.deleteConnectionsAndVars("connection_id", id)
	})
}

func (gdb *gormdb) updateConnection(ctx context.Context, id sdktypes.UUID, data map[string]any) error {
	return gdb.transaction(ctx, func(tx *tx) error {
		if err := tx.isCtxUserEntity(ctx, id); err != nil {
			return err
		}
		return tx.db.Model(&scheme.Connection{ConnectionID: id}).Updates(data).Error
	})
}

func (gdb *gormdb) getConnection(ctx context.Context, id sdktypes.UUID) (*scheme.Connection, error) {
	return getOne[scheme.Connection](gdb.withUserConnections(ctx), "connection_id = ?", id)
}

func findConnections(query *gorm.DB) ([]scheme.Connection, error) {
	var cs []scheme.Connection
	if err := query.Find(&cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

func (gdb *gormdb) getConnections(ctx context.Context, ids ...sdktypes.UUID) ([]scheme.Connection, error) {
	q := gdb.withUserConnections(ctx).Where("connection_id IN (?)", ids)
	return findConnections(q)
}

func (gdb *gormdb) listConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter, idsOnly bool) ([]scheme.Connection, error) {
	q := gdb.withUserConnections(ctx)

	if filter.IntegrationID.IsValid() {
		q = q.Where("integration_id = ?", filter.IntegrationID.UUIDValue())
	}

	if filter.ProjectID.IsValid() {
		q = q.Where("project_id = ?", filter.ProjectID.UUIDValue())
	}

	if filter.StatusCode != sdktypes.StatusCodeUnspecified {
		q = q.Where("status_code = ?", int32(filter.StatusCode.ToProto()))
	}

	if idsOnly {
		q = q.Select("connection_id")
	}
	return findConnections(q)
}

func (db *gormdb) CreateConnection(ctx context.Context, conn sdktypes.Connection) error {
	if err := conn.Strict(); err != nil {
		return err
	}

	c := scheme.Connection{
		ConnectionID:  conn.ID().UUIDValue(),
		IntegrationID: scheme.UUIDOrNil(conn.IntegrationID().UUIDValue()),
		ProjectID:     scheme.UUIDOrNil(conn.ProjectID().UUIDValue()),
		Name:          conn.Name().String(),
		StatusCode:    int32(conn.Status().Code().ToProto()),
		StatusMessage: conn.Status().Message(),
	}

	return translateError(db.createConnection(ctx, &c))
}

func (db *gormdb) DeleteConnection(ctx context.Context, id sdktypes.ConnectionID) error {
	return translateError(db.deleteConnection(ctx, id.UUIDValue()))
}

func (db *gormdb) UpdateConnection(ctx context.Context, conn sdktypes.Connection) error {
	// This will never update integration id or project id, so not checking them.

	data := make(map[string]any, 2)

	if conn.Name().IsValid() {
		data["name"] = conn.Name().String()
	}

	if conn.Status().IsValid() {
		data["status_code"] = int32(conn.Status().Code().ToProto())
		data["status_message"] = conn.Status().Message()
	}
	if len(data) == 0 {
		return nil
	}
	return translateError(db.updateConnection(ctx, conn.ID().UUIDValue(), data))
}

func (db *gormdb) GetConnection(ctx context.Context, connectionID sdktypes.ConnectionID) (sdktypes.Connection, error) {
	c, err := db.getConnection(ctx, connectionID.UUIDValue())
	if c == nil || err != nil {
		return sdktypes.InvalidConnection, translateError(err)
	}
	return scheme.ParseConnection(*c)
}

func (db *gormdb) GetConnections(ctx context.Context, connectionIDs []sdktypes.ConnectionID) ([]sdktypes.Connection, error) {
	ids := kittehs.Transform(connectionIDs, func(id sdktypes.ConnectionID) uuid.UUID { return id.UUIDValue() })
	cs, err := db.getConnections(ctx, ids...)
	if err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(cs, scheme.ParseConnection)
}

func (db *gormdb) ListConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter, idsOnly bool) ([]sdktypes.Connection, error) {
	cs, err := db.listConnections(ctx, filter, idsOnly)
	if err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(cs, scheme.ParseConnection)
}
