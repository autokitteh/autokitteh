package triggers

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/db"
	"go.autokitteh.dev/autokitteh/internal/backend/dispatcher"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type triggers struct {
	z      *zap.Logger
	db     db.DB
	tmprl  client.Client
	events sdkservices.Events
}

func New(z *zap.Logger, db db.DB, t client.Client, e sdkservices.Events) sdkservices.Triggers {
	return &triggers{db: db, z: z, tmprl: t, events: e}
}

func (m *triggers) Create(ctx context.Context, trigger sdktypes.Trigger) (sdktypes.TriggerID, error) {
	if trigger.ID().IsValid() {
		return sdktypes.InvalidTriggerID, errors.New("trigger id already defined")
	}
	trigger = trigger.WithNewID()

	if schedule, found := trigger.Data()["schedule"]; found && schedule.IsValid() {
		return m.createScheduledTrigger(ctx, trigger, schedule)
	}

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
	trigger, err := m.db.GetTrigger(ctx, triggerID)
	if err != nil {
		return err
	}

	data := trigger.Data()
	if schedule, found := data["schedule"]; found && schedule.IsValid() {
		return m.deleteScheduledTrigger(ctx, triggerID, data)
	}
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

func (m *triggers) createScheduledTrigger(ctx context.Context, trigger sdktypes.Trigger, schedule sdktypes.Value) (sdktypes.TriggerID, error) {
	scheduleID, err := m.createScheduledWorkflow(ctx, schedule, trigger)
	if err != nil {
		return sdktypes.InvalidTriggerID, fmt.Errorf("create scheduler trigger: %w", err)
	}

	trigger = trigger.WithUpdatedData("schedule_id", sdktypes.NewStringValue(*scheduleID))

	if err := m.db.CreateTrigger(ctx, trigger); err != nil {
		return sdktypes.InvalidTriggerID, err
	}

	return trigger.ID(), nil
}

func (m *triggers) deleteScheduledTrigger(ctx context.Context, triggerID sdktypes.TriggerID, data map[string]sdktypes.Value) error {
	scheduleIDVal, found := data["schedule_id"]
	scheduleID, err := scheduleIDVal.ToString()

	if !found || err != nil {
		m.z.Error("Failed delete scheduler workflow. No schedulerID found")
		err = fmt.Errorf("delete trigger: scheduleID not found")
	}

	if err == nil {
		scheduleHandle := m.tmprl.ScheduleClient().GetHandle(ctx, scheduleID)
		err = scheduleHandle.Delete(ctx)
	}

	// TODO: create and save cancel event?
	return errors.Join(err, m.db.DeleteTrigger(ctx, triggerID))
}

func (m *triggers) createScheduledWorkflow(ctx context.Context, schedule sdktypes.Value, trigger sdktypes.Trigger) (*string, error) {
	scheduleStr, err := schedule.ToString()
	if err != nil || scheduleStr == "" {
		return nil, fmt.Errorf("failed to parse trigger schedule <%v>: %w", schedule, err)
	}

	event := kittehs.Must1(sdktypes.EventFromProto(&sdktypes.EventPB{EventType: "scheduler"}))
	eventID, err := m.events.Save(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("save event: %w", err)
	}

	z := m.z.With(zap.String("event_id", eventID.String()))
	z.Debug("create scheduled event", zap.String("schedule", scheduleStr))

	scheduleID := fmt.Sprintf("%s_%s", trigger.Name().String(), trigger.ID().String())
	triggerID := trigger.ID()
	_, err = m.tmprl.ScheduleClient().Create(ctx,
		client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{scheduleStr},
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        eventID.String(), // workflowID
				Workflow:  "events_workflow",
				TaskQueue: "events-task-queue",
				Args: []interface{}{
					dispatcher.EventsWorkflowInput{EventID: eventID, TriggerID: &triggerID},
				},
			},
		})
	if err != nil {
		m.z.Error("Failed starting scheduler workflow, orphaned event", zap.Error(err))
		return nil, fmt.Errorf("failed to create trigger schedule: %w", err)
	}
	z.Debug("created scheduled workflow", zap.String("schedule_id", scheduleID))
	return &scheduleID, nil
}
