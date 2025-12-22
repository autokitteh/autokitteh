package dbgorm

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"maps"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) createConnection(ctx context.Context, conn *scheme.Connection) error {
	return gdb.writeTransaction(ctx, func(tx *gormdb) error {
		if gdb.cfg.Type == "postgres" {
			h := fnv.New64a()
			h.Write(conn.OrgID[:])
			h.Write([]byte(conn.Name))
			lockID := int64(h.Sum64())

			if err := tx.writer.Exec("SELECT pg_advisory_xact_lock(?)", lockID).Error; err != nil {
				return err
			}
		}

		var conflicts []scheme.Connection
		query := tx.writer.Where("org_id = ? AND name = ? AND deleted_at IS NULL", conn.OrgID, conn.Name)
		if err := query.Find(&conflicts).Error; err != nil {
			return err
		}

		if conn.ProjectID == nil {
			if len(conflicts) > 0 {
				return errors.New("connection name already in use in this organization")
			}
		} else {
			for _, conflict := range conflicts {
				if conflict.ProjectID == nil {
					return errors.New("connection name already used at organization level")
				}
				if *conflict.ProjectID == *conn.ProjectID {
					return errors.New("connection name already exists in this project")
				}
			}
		}
		return tx.writer.Create(conn).Error
	})
}

func (gdb *gormdb) deleteConnectionsAndVars(ctx context.Context, what string, id uuid.UUID) error {
	// should be transactional with context already applied
	if what == "connection_id" {
		hasTriggers, err := gdb.doesConnectionHaveTriggers(ctx, id)
		if err != nil {
			return translateError(err)
		}
		if hasTriggers {
			return errors.New("cannot delete a connection that has associated triggers")
		}
	}

	var ids []uuid.UUID
	q := gdb.writer.WithContext(ctx).Model(&scheme.Connection{})
	q = q.Clauses(clause.Returning{Columns: []clause.Column{{Name: "connection_id"}}})
	if err := q.Delete(&ids, what+" = ?", id).Error; err != nil {
		return err
		// REVIEW: proceed to vars deletion if there are any?
	}

	if len(ids) > 0 {
		return gdb.writer.Where("var_id IN (?)", ids).Delete(&scheme.Var{}).Error
	}

	return nil
}

func (gdb *gormdb) doesConnectionHaveTriggers(ctx context.Context, connID uuid.UUID) (bool, error) {
	var exists bool
	q := gdb.reader.WithContext(ctx)

	q = q.Model(&scheme.Trigger{}).
		Select("1").
		Where("connection_id = ?", connID).
		Where("deleted_at IS NULL").
		Limit(1)

	err := q.Find(&exists).Error
	if err != nil {
		return false, fmt.Errorf("checking active triggers: %w", err)
	}
	return exists, nil
}

func (gdb *gormdb) deleteConnection(ctx context.Context, id uuid.UUID) error {
	return gdb.deleteConnectionsAndVars(ctx, "connection_id", id)
}

func (gdb *gormdb) updateConnection(ctx context.Context, id uuid.UUID, data map[string]any) error {
	if _, exists := data["name"]; !exists {
		return gdb.writer.WithContext(ctx).Model(&scheme.Connection{ConnectionID: id}).Updates(data).Error
	}

	return gdb.writeTransaction(ctx, func(tx *gormdb) error {
		conn, err := gdb.getConnection(ctx, id)
		if err != nil {
			return err
		}

		if gdb.cfg.Type == "postgres" {
			h := fnv.New64a()
			h.Write(conn.OrgID[:])
			h.Write([]byte(conn.Name))
			lockID := int64(h.Sum64())

			if err := tx.writer.Exec("SELECT pg_advisory_xact_lock(?)", lockID).Error; err != nil {
				return err
			}
			// need to lock the new name as well
			// to prevent edit and creation at the same time
			h2 := fnv.New64a()
			h2.Write(conn.OrgID[:])
			h2.Write([]byte(data["name"].(string)))
			lockID2 := int64(h2.Sum64())

			if err := tx.writer.Exec("SELECT pg_advisory_xact_lock(?)", lockID2).Error; err != nil {
				return err
			}
		}

		var connectionsWithSameName []scheme.Connection
		query := tx.writer.Where("org_id = ? AND name = ? AND deleted_at IS NULL", conn.OrgID, data["name"])
		if err := query.Find(&connectionsWithSameName).Error; err != nil {
			return err
		}

		if conn.ProjectID == nil {
			// this connection is org connection, can't have any other conflict
			if len(connectionsWithSameName) > 0 {
				return errors.New("duplicate name")
			}
		} else {
			// project level connection
			for _, otherConnectionWithSamename := range connectionsWithSameName {
				if otherConnectionWithSamename.ProjectID == conn.ProjectID {
					return errors.New("duplicate name in same project")
				}
			}
		}
		return gdb.writer.WithContext(ctx).Model(&scheme.Connection{ConnectionID: id}).Updates(data).Error
	})
}

func (gdb *gormdb) getConnection(ctx context.Context, id uuid.UUID) (*scheme.Connection, error) {
	return getOne[scheme.Connection](gdb.reader.WithContext(ctx), "connection_id = ?", id)
}

func (gdb *gormdb) getConnections(ctx context.Context, ids ...uuid.UUID) ([]scheme.Connection, error) {
	var cs []scheme.Connection
	if err := gdb.reader.WithContext(ctx).Where("connection_id IN (?)", ids).Find(&cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

func (gdb *gormdb) listConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter, idsOnly bool) ([]scheme.Connection, error) {
	q := gdb.reader.WithContext(ctx)

	if filter.OrgID.IsValid() {
		x := q.Where("org_id = ? AND project_id IS NULL", filter.OrgID.UUIDValue())
		if filter.ProjectID.IsValid() {
			x = x.Or(q.Where("org_id = ? AND project_id = ?", filter.OrgID.UUIDValue(), filter.ProjectID.UUIDValue()))
		}
		q = x
	} else if filter.ProjectID.IsValid() {
		q = q.Where("project_id = ?", filter.ProjectID.UUIDValue())
	}

	if filter.IntegrationID.IsValid() {
		q = q.Where("integration_id = ?", filter.IntegrationID.UUIDValue())
	}

	if filter.StatusCode != sdktypes.StatusCodeUnspecified {
		q = q.Where("status_code = ?", int32(filter.StatusCode.ToProto()))
	}

	if idsOnly {
		q = q.Select("connection_id")
	}

	var cs []scheme.Connection
	if err := q.Find(&cs).Error; err != nil {
		return nil, err
	}
	return cs, nil
}

func (db *gormdb) CreateConnection(ctx context.Context, conn sdktypes.Connection) error {
	if err := conn.Strict(); err != nil {
		return err
	}

	if !conn.OrgID().IsValid() {
		return errors.New("org ID is required")
	}

	c := scheme.Connection{
		Base:          based(ctx),
		ProjectID:     conn.ProjectID().UUIDValuePtr(),
		OrgID:         conn.OrgID().UUIDValue(),
		ConnectionID:  conn.ID().UUIDValue(),
		IntegrationID: uuidPtrOrNil(conn.IntegrationID()),
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

	maps.Copy(data, updatedBaseColumns(ctx))

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
