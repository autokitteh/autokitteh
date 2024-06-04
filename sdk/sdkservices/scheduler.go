package sdkservices

import (
	"context"

	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Scheduler interface {
	Create(ctx context.Context, scheduleID string, schedule string, triggerID sdktypes.TriggerID) error
	Delete(ctx context.Context, scheduleID string) error
	Update(ctx context.Context, scheduleID string, schedule string) error
}
