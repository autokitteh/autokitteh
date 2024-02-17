package mappings

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type mappings struct {
	z  *zap.Logger
	db db.DB
}

func New(z *zap.Logger, db db.DB) sdkservices.Mappings {
	return &mappings{db: db, z: z}
}

func (m *mappings) Create(ctx context.Context, mapping sdktypes.Mapping) (sdktypes.MappingID, error) {
	if mid := sdktypes.GetMappingID(mapping); mid != nil {
		return nil, errors.New("mapping id already defined")
	}

	mapping, err := mapping.Update(func(pb *sdktypes.MappingPB) {
		pb.MappingId = sdktypes.NewMappingID().String()
	})
	if err != nil {
		return nil, err
	}

	if err = m.db.CreateMapping(ctx, mapping); err != nil {
		return nil, err
	}

	return sdktypes.GetMappingID(mapping), nil
}

// Delete implements sdkservices.Mappings.
func (m *mappings) Delete(ctx context.Context, mappingID sdktypes.MappingID) error {
	return m.db.DeleteMapping(ctx, mappingID)
}

// Get implements sdkservices.Mappings.
func (m *mappings) Get(ctx context.Context, mappingID sdktypes.MappingID) (sdktypes.Mapping, error) {
	return sdkerrors.IgnoreNotFoundErr(m.db.GetMapping(ctx, mappingID))
}

// List implements sdkservices.Mappings.
func (m *mappings) List(ctx context.Context, filter sdkservices.ListMappingsFilter) ([]sdktypes.Mapping, error) {
	return m.db.ListMappings(ctx, filter)
}
