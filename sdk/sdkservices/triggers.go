package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type ListTriggersFilter struct {
	ProjectID    sdktypes.ProjectID
	ConnectionID sdktypes.ConnectionID
	SourceType   sdktypes.TriggerSourceType
}

type Triggers interface {
	Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error)
	Update(ctx context.Context, trigger sdktypes.Trigger) error
	Delete(ctx context.Context, triggerID sdktypes.TriggerID) error
	Get(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error)
	List(ctx context.Context, filter ListTriggersFilter) ([]sdktypes.Trigger, error)
}
