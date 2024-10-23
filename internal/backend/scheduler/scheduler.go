package scheduler

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	taskQueueName = "scheduler-task-queue"
	workerID      = "scheduler-worker"
	workflowName  = "scheduler_workflow"
)

type Scheduler struct {
	temporal   temporalclient.Client
	sl         *zap.SugaredLogger
	dispatcher sdkservices.Dispatcher
	triggers   sdkservices.Triggers
}

func New(l *zap.Logger, tc temporalclient.Client) *Scheduler {
	return &Scheduler{sl: l.Sugar(), temporal: tc}
}

func (sch *Scheduler) Start(ctx context.Context, dispatcher sdkservices.Dispatcher, triggers sdkservices.Triggers) error {
	sch.dispatcher = dispatcher
	sch.triggers = triggers

	w := worker.New(sch.temporal.Temporal(), taskQueueName, worker.Options{Identity: workerID})
	w.RegisterWorkflowWithOptions(sch.workflow, workflow.RegisterOptions{Name: workflowName})

	if err := w.Start(); err != nil {
		return fmt.Errorf("schedule wf: worker start: %w", err)
	}

	return nil
}

func (sch *Scheduler) Create(ctx context.Context, tid sdktypes.TriggerID, schedule string) error {
	l := sch.sl.With("trigger_id", tid.String())

	_, err := sch.temporal.Temporal().ScheduleClient().Create(
		ctx,
		client.ScheduleOptions{
			ID: tid.String(),
			Spec: client.ScheduleSpec{
				CronExpressions: []string{schedule},
			},
			Action: &client.ScheduleWorkflowAction{
				ID:        tid.String(), // workflowID
				Workflow:  workflowName,
				TaskQueue: taskQueueName,
				Args:      []any{tid},
			},
			Memo: map[string]any{
				"trigger_id":   tid.String(),
				"trigger_uuid": tid.UUIDValue().String(),
			},
		},
	)
	if err != nil {
		return fmt.Errorf("schedule: create schedule workflow: %w", err)
	}

	l.With("schedule", schedule).Infof("created schedule %q for %v", schedule, tid)

	return nil
}

func (sch *Scheduler) Delete(ctx context.Context, tid sdktypes.TriggerID) error {
	sl := sch.sl.With("trigger_id", tid)

	scheduleHandle := sch.temporal.Temporal().ScheduleClient().GetHandle(ctx, tid.String()) // validity of scheduleID is not checked by temporal
	if err := scheduleHandle.Delete(ctx); err != nil {
		return fmt.Errorf("schedule: delete scheduler workflow: %w", err)
	}

	sl.Infof("deleted schedule workflow for %s", tid)

	return nil
}

func (sch *Scheduler) Update(ctx context.Context, tid sdktypes.TriggerID, schedule string) error {
	sl := sch.sl.With("trigger_id", tid)

	h := sch.temporal.Temporal().ScheduleClient().GetHandle(ctx, tid.String()) // validity of scheduleID is not checked by temporal
	err := h.Update(ctx, client.ScheduleUpdateOptions{
		DoUpdate: func(input client.ScheduleUpdateInput) (*client.ScheduleUpdate, error) {
			input.Description.Schedule.Spec = &client.ScheduleSpec{CronExpressions: []string{schedule}}
			return &client.ScheduleUpdate{Schedule: &input.Description.Schedule}, nil
		},
	})
	if err != nil {
		return fmt.Errorf("schedule: update scheduler workflow: %w", err)
	}

	sl.With("schedule", schedule).Infof("updated schedule %v to schedule %q", tid, schedule)

	return nil
}

func (sch *Scheduler) workflow(wctx workflow.Context, tid sdktypes.TriggerID) error {
	sl := sch.sl.With("trigger_id", tid)

	ctx := temporalclient.NewWorkflowContextAsGOContext(wctx)

	// It's ok to run all these straight in the workflow - they're all should be
	// pretty fast.

	t, err := sch.triggers.Get(ctx, tid)
	if err != nil {
		if !errors.Is(err, sdkerrors.ErrNotFound) {
			return fmt.Errorf("get trigger: %w", err)
		}
	}

	if !t.IsValid() {
		sl.Warnf("trigger %v not found, removing schedule", tid)

		if err := sch.Delete(ctx, tid); err != nil {
			return fmt.Errorf("delete schedule: %w", err)
		}

		return nil
	}

	eid, err := sch.dispatcher.Dispatch(ctx, sdktypes.NewEvent(tid).WithType("tick"), nil)
	if err != nil {
		return fmt.Errorf("dispatch event: %w", err)
	}

	sl.With("event_id", eid).Infof("schedule event workflow for %v dispatched as %v", tid, eid)

	return nil
}
