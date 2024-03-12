package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) createConnection(ctx context.Context, conn scheme.Connection) error {
	return db.db.WithContext(ctx).Create(&conn).Error
}

func (db *gormdb) CreateConnection(ctx context.Context, conn sdktypes.Connection) error {
	c := scheme.Connection{
		ConnectionID:     conn.ID().String(),
		IntegrationID:    conn.IntegrationID().String(), // TODO(ENG-158): need to verify integration id
		IntegrationToken: conn.IntegrationToken(),
		ProjectID:        conn.ProjectID().String(), // TODO(ENG-136): need to verify parent id
		Name:             conn.Name().String(),
	}

	return translateError(db.createConnection(ctx, c))
}

func (db *gormdb) UpdateConnection(ctx context.Context, conn sdktypes.Connection) error {
	c := scheme.Connection{
		ConnectionID:     conn.ID().String(),
		IntegrationID:    conn.IntegrationID().String(), // TODO(ENG-158): need to verify integration id
		IntegrationToken: conn.IntegrationToken(),
		ProjectID:        conn.ProjectID().String(), // TODO(ENG-136): need to verify parent id
		Name:             conn.Name().String(),
	}

	err := db.db.WithContext(ctx).
		Where("connection_id = ?", conn.ID().String()).
		Updates(&c).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) deleteConnection(ctx context.Context, id string) error {
	return db.db.WithContext(ctx).Delete(&scheme.Connection{ConnectionID: id}).Error
}

func (db *gormdb) DeleteConnection(ctx context.Context, id sdktypes.ConnectionID) error {
	return translateError(db.deleteConnection(ctx, id.String()))
}

func (db *gormdb) GetConnection(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseConnection, "connection_id = ?", id.String())
}

func (db *gormdb) ListConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	q := db.db.WithContext(ctx)
	if filter.IntegrationID.IsValid() {
		q = q.Where("integration_id = ?", filter.IntegrationID.String())
	}
	if filter.IntegrationToken != "" {
		q = q.Where("integration_token = ?", filter.IntegrationToken)
	}
	if filter.ProjectID.IsValid() {
		q = q.Where("project_id = ?", filter.ProjectID.String())
	}

	var cs []scheme.Connection
	if err := q.Find(&cs).Error; err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(cs, scheme.ParseConnection)
}
