package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) createConnection(ctx context.Context, conn *scheme.Connection) error {
	return db.db.WithContext(ctx).Create(conn).Error
}

func (db *gormdb) CreateConnection(ctx context.Context, conn sdktypes.Connection) error {
	c := scheme.Connection{
		ConnectionID:     *conn.ID().Value(),
		IntegrationID:    conn.ProjectID().Value(),
		IntegrationToken: conn.IntegrationToken(),
		ProjectID:        conn.ProjectID().Value(),
		Name:             conn.Name().String(),
	}

	return translateError(db.createConnection(ctx, &c))
}

func (db *gormdb) UpdateConnection(ctx context.Context, conn sdktypes.Connection) error {
	c := scheme.Connection{
		ConnectionID:     *conn.ID().Value(),
		IntegrationID:    conn.IntegrationID().Value(),
		IntegrationToken: conn.IntegrationToken(),
		ProjectID:        conn.ProjectID().Value(),
		Name:             conn.Name().String(),
	}

	err := db.db.WithContext(ctx).
		Where("connection_id = ?", conn.ID().Value()).
		Updates(&c).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) deleteConnection(ctx context.Context, id sdktypes.UUID) error {
	return db.db.WithContext(ctx).Delete(&scheme.Connection{ConnectionID: id}).Error
}

func (db *gormdb) DeleteConnection(ctx context.Context, id sdktypes.ConnectionID) error {
	return translateError(db.deleteConnection(ctx, *id.Value()))
}

func (db *gormdb) GetConnection(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseConnection, "connection_id = ?", id.Value())
}

func (db *gormdb) ListConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	q := db.db.WithContext(ctx)
	if filter.IntegrationID.IsValid() {
		q = q.Where("integration_id = ?", filter.IntegrationID.Value())
	}
	if filter.IntegrationToken != "" {
		q = q.Where("integration_token = ?", filter.IntegrationToken)
	}
	if filter.ProjectID.IsValid() {
		q = q.Where("project_id = ?", filter.ProjectID.Value())
	}

	var cs []scheme.Connection
	if err := q.Find(&cs).Error; err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(cs, scheme.ParseConnection)
}
