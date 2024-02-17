package dbgorm

import (
	"context"

	"gorm.io/gorm/clause"

	"go.autokitteh.dev/autokitteh/backend/internal/db/dbgorm/scheme"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (db *gormdb) CreateMapping(ctx context.Context, mapping sdktypes.Mapping) error {
	mappingEvents := kittehs.Transform(sdktypes.GetMappingEvents(mapping), func(e sdktypes.MappingEvent) scheme.MappingEvent {
		loc := sdktypes.GetMappingEventCodeLocation(e)
		return scheme.MappingEvent{
			Type:         sdktypes.GetMappingEventType(e),
			CodeLocation: sdktypes.GetCodeLocationCanonicalString(loc),
		}
	})
	m := scheme.Mapping{
		MappingID:    sdktypes.GetMappingID(mapping).String(),
		EnvID:        sdktypes.GetMappingEnvID(mapping).String(),
		ConnectionID: sdktypes.GetMappingConnectionID(mapping).String(),
		ModuleName:   sdktypes.GetMappingModuleName(mapping).String(),
		Events:       mappingEvents,
	}

	// Assuming if eventID already exists, nothing will happen and no error
	if err := db.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&m).Error; err != nil {
		return translateError(err)
	}
	return nil
}

func (db *gormdb) GetMapping(ctx context.Context, mid sdktypes.MappingID) (sdktypes.Mapping, error) {
	return get(db.db.Preload("Events"), ctx, scheme.ParseMapping, "mapping_id = ?", mid.String())
}

func (db *gormdb) DeleteMapping(ctx context.Context, mid sdktypes.MappingID) error {
	var m scheme.Mapping
	if err := db.db.WithContext(ctx).Where("mapping_id = ?", mid.String()).Delete(&m).Error; err != nil {
		return translateError(err)
	}

	return nil
}

func (db *gormdb) ListMappings(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
	q := db.db.WithContext(ctx)
	if filter.EnvID != nil {
		q = q.Where("env_id = ?", filter.EnvID.String())
	}

	if filter.ConnectionID != nil {
		q = q.Where("connection_id = ?", filter.ConnectionID.String())
	}

	var es []scheme.Mapping
	if err := q.Preload("Events").Find(&es).Error; err != nil {
		return nil, translateError(err)
	}

	return kittehs.TransformError(es, scheme.ParseMapping)
}
