package schedule

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	wf "go.autokitteh.dev/autokitteh/internal/backend/workflows"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	SchedulerEventTriggerType = "scheduler"
	ScheduleExpression        = "schedule"
	ScheduleIDKey             = "schedule_id"
	SchedulerConnectionName   = "cron"
)

type temporalSchedule struct {
	tmprl  temporalclient.Client
	events sdkservices.Events
	z      *zap.Logger
}

func New(z *zap.Logger, events sdkservices.Events, tc temporalclient.Client) sdkservices.Scheduler {
	return &temporalSchedule{z: z, events: events, tmprl: tc}
}

func (tsc *temporalSchedule) CreateEventRecord(ctx context.Context, eventID sdktypes.EventID, state sdktypes.EventState) {
	record := sdktypes.NewEventRecord(eventID, state)
	if err := tsc.events.AddEventRecord(ctx, record); err != nil {
		tsc.z.Panic("Failed setting event state", zap.String("eventID", eventID.String()), zap.String("state", state.String()), zap.Error(err))
	}
}

func (tsc *temporalSchedule) Create(ctx context.Context, scheduleID string, schedule string, triggerID sdktypes.TriggerID) error {
	z := tsc.z.With(zap.String("trigger_id", triggerID.String())).With(zap.String("schedule_id", scheduleID))
	z.Debug("create schedule event", zap.String("schedule", schedule))

	_, err := tsc.tmprl.Temporal().ScheduleClient().Create(ctx,
		client.ScheduleOptions{
			ID: scheduleID,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{schedule},
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        triggerID.String(), // workflowID
				Workflow:  wf.SchedulerWorkflow,
				TaskQueue: wf.ScheduleTaskQueueName,
				Args:      []any{triggerID},
			},
		})
	if err != nil {
		z.Error("Failed creating schedule workflow", zap.Error(err))
		return fmt.Errorf("schedule: create schedule workflow: %w", err)
	}
	z.Info("created schedule workflow")
	return nil
}

func (tsc *temporalSchedule) Delete(ctx context.Context, scheduleID string) error {
	if scheduleID == "" {
		tsc.z.Error("Failed deleting schedule workflow. No scheduleID found")
		return fmt.Errorf("schedule: delete schedule workflow: no scheduleID")
	}
	z := tsc.z.With(zap.String("schedule_id", scheduleID))

	scheduleHandle := tsc.tmprl.Temporal().ScheduleClient().GetHandle(ctx, scheduleID) // validity of scheduleID is not checked by temporal
	if err := scheduleHandle.Delete(ctx); err != nil {
		z.Error("Failed deleting schedule workflow", zap.Error(err))
		return fmt.Errorf("schedule: delete scheduler workflow: %w", err)
	}
	z.Info("deleted schedule workflow")
	// REVIEW: create and save cancel schedule event?
	return nil
}

func (tsc *temporalSchedule) Update(ctx context.Context, scheduleID string, scheduleStr string) error {
	if scheduleID == "" {
		tsc.z.Error("Failed updating schedule workflow. No scheduleID found")
		return fmt.Errorf("schedule: upadate schedule workflow: no scheduleID")
	}
	z := tsc.z.With(zap.String("schedule_id", scheduleID))

	scheduleHandle := tsc.tmprl.Temporal().ScheduleClient().GetHandle(ctx, scheduleID) // validity of scheduleID is not checked by temporal
	err := scheduleHandle.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(schedule client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			schedule.Description.Schedule.Spec = &client.ScheduleSpec{CronExpressions: []string{scheduleStr}}
			return &client.ScheduleUpdate{Schedule: &schedule.Description.Schedule}, nil
		},
	})
	if err != nil {
		z.Error("Failed updating schedule workflow", zap.Error(err))
		return fmt.Errorf("schedule: update scheduler workflow: %w", err)
	}
	z.Info("updated schedule workflow")
	return nil
}
