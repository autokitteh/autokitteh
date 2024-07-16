package dbgorm

import (
	"context"
	"encoding/json"
	"net/url"

	"go.autokitteh.dev/autokitteh/internal/backend/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (gdb *gormdb) createIntegration(ctx context.Context, i *scheme.Integration) error {
	return gdb.db.WithContext(ctx).Create(i).Error
}

func (gdb *gormdb) deleteIntegration(ctx context.Context, id sdktypes.UUID) error {
	return gdb.db.WithContext(ctx).Delete(&scheme.Integration{IntegrationID: id}).Error
}

func (gdb *gormdb) listIntegrations(ctx context.Context) ([]scheme.Integration, error) {
	var is []scheme.Integration
	if err := gdb.db.WithContext(ctx).Find(&is).Error; err != nil {
		return nil, err
	}
	return is, nil
}

func convertTypeToRecord(i sdktypes.Integration) *scheme.Integration {
	l := i.LogoURL()
	if l == nil {
		l = &url.URL{}
	}

	uls, err := json.Marshal(i.UserLinks())
	if err != nil {
		uls = []byte{}
	}

	c := i.ConnectionURL()
	if c == nil {
		c = &url.URL{}
	}

	return &scheme.Integration{
		IntegrationID: i.ID().UUIDValue(),
		UniqueName:    i.UniqueName().String(),
		DisplayName:   i.DisplayName(),
		Description:   i.Description(),
		LogoURL:       l.String(),
		UserLinks:     uls,
		// TODO: Tags
		// TODO(ENG-346): Connection UI specification instead of a URL.
		ConnectionURL: c.String(),
		// TODO: Functions
		// TODO: Events
		// TODO: APIKey
		// TODO: SigningKey
	}
}

func (db *gormdb) CreateIntegration(ctx context.Context, integration sdktypes.Integration) error {
	if err := integration.Strict(); err != nil {
		return err
	}
	i := convertTypeToRecord(integration)
	return translateError(db.createIntegration(ctx, i))
}

func (db *gormdb) DeleteIntegration(ctx context.Context, id sdktypes.IntegrationID) error {
	// Desired product behavior: if user tries to delete an integration which
	// already has associated connections, AK should confirm with the user
	// what they want to do - abort, or cascade the deletion.
	// Note that a similar decision exists when deleting connections that
	// have active project mappings.
	return translateError(db.deleteIntegration(ctx, id.UUIDValue()))
}

func (db *gormdb) UpdateIntegration(ctx context.Context, i sdktypes.Integration) error {
	integ := convertTypeToRecord(i)
	err := db.db.WithContext(ctx).
		Where("integration_id = ?", integ.IntegrationID).
		Updates(integ).Error
	if err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) GetIntegration(ctx context.Context, id sdktypes.IntegrationID) (sdktypes.Integration, error) {
	return getOneWTransform(db.db, ctx, scheme.ParseIntegration, "integration_id = ?", id.UUIDValue())
}

func (db *gormdb) ListIntegrations(ctx context.Context) ([]sdktypes.Integration, error) {
	is, err := db.listIntegrations(ctx)
	if is == nil || err != nil {
		return nil, translateError(err)
	}
	return kittehs.TransformError(is, scheme.ParseIntegration)
}
