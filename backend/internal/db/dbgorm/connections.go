package dbgorm

import (
	"context"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) CreateConnection(ctx context.Context, conn sdktypes.Connection) error {
	c := scheme.Connection{
		ConnectionID:     sdktypes.GetConnectionID(conn).String(),
		IntegrationID:    sdktypes.GetConnectionIntegrationID(conn).String(), // TODO(ENG-158): need to verify integration id
		IntegrationToken: sdktypes.GetConnectionIntegrationToken(conn),
		ProjectID:        sdktypes.GetConnectionProjectID(conn).String(), // TODO(ENG-136): need to verify parent id
		Name:             sdktypes.GetConnectionName(conn).String(),
	}

	if err := db.db.WithContext(ctx).Create(&c).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) UpdateConnection(ctx context.Context, conn sdktypes.Connection) error {
	c := scheme.Connection{
		ConnectionID:     sdktypes.GetConnectionID(conn).String(),
		IntegrationID:    sdktypes.GetConnectionIntegrationID(conn).String(), // TODO(ENG-158): verify integration id
		IntegrationToken: sdktypes.GetConnectionIntegrationToken(conn),
		ProjectID:        sdktypes.GetConnectionProjectID(conn).String(), // TODO(ENG-136): need to verify parent id
		Name:             sdktypes.GetConnectionName(conn).String(),
	}

	err := db.db.WithContext(ctx).
		Where("connection_id = ?", sdktypes.GetConnectionID(conn).String()).
		Updates(&c).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) DeleteConnection(ctx context.Context, id sdktypes.ConnectionID) error {
	var c scheme.Connection
	err := db.db.WithContext(ctx).
		Where("connection_id = ?", id.String()).
		Delete(&c).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) GetConnection(ctx context.Context, id sdktypes.ConnectionID) (sdktypes.Connection, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseConnection, "connection_id = ?", id.String())
}

func (db *gormdb) ListConnections(ctx context.Context, filter sdkservices.ListConnectionsFilter) ([]sdktypes.Connection, error) {
	q := db.db.WithContext(ctx)
	if filter.IntegrationID != nil {
		q = q.Where("integration_id = ?", filter.IntegrationID.String())
	}
	if filter.IntegrationToken != "" {
		q = q.Where("integration_token = ?", filter.IntegrationToken)
	}
	if filter.ProjectID != nil {
		q = q.Where("project_id = ?", filter.ProjectID.String())
	}

	var cs []scheme.Connection
	if err := q.Find(&cs).Error; err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(cs, scheme.ParseConnection)
}
