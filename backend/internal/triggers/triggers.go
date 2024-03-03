package triggers

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/backend/internal/db"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type triggers struct {
	z  *zap.Logger
	db db.DB
}

func New(z *zap.Logger, db db.DB) sdkservices.Triggers {
	return &triggers{db: db, z: z}
}

func (m *triggers) Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error) {
	if trigger.ID().IsValid() {
		return sdktypes.InvalidTriggerID, errors.New("trigger id already defined")
	}

	trigger = trigger.WithNewID()

	if err := m.db.CreateTrigger(ctx, trigger); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	return trigger.ID(), nil
}

func (m *triggers) Update(ctx context.Context, trigger sdktypes.Trigger) error {
	return m.db.UpdateTrigger(ctx, trigger)
}

// Delete implements sdkservices.Triggers.
func (m *triggers) Delete(ctx context.Context, triggerID sdktypes.TriggerID) error {
	return m.db.DeleteTrigger(ctx, triggerID)
}

// Get implements sdkservices.Triggers.
func (m *triggers) Get(ctx context.Context, triggerID sdktypes.TriggerID) (sdktypes.Trigger, error) {
	return sdkerrors.IgnoreNotFoundErr(m.db.GetTrigger(ctx, triggerID))
}

// List implements sdkservices.Triggers.
func (m *triggers) List(ctx context.Context, filter sdkservices.ListTriggersFilter) ([]sdktypes.Trigger, error) {
	return m.db.ListTriggers(ctx, filter)
}
