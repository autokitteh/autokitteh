package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListMappingsFilter struct {
	EnvID        sdktypes.EnvID
	ConnectionID sdktypes.ConnectionID
}

type Mappings interface {
	Create(ctx context.Context, mapping sdktypes.Mapping) (sdktypes.MappingID, error)
	Delete(ctx context.Context, mappingID sdktypes.MappingID) error
	Get(ctx context.Context, mappingID sdktypes.MappingID) (sdktypes.Mapping, error)
	List(ctx context.Context, filter ListMappingsFilter) ([]sdktypes.Mapping, error)
}
