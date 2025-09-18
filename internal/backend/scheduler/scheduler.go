package scheduler

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/backend/temporalclient"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

const (
	taskQueueName = "scheduler-task-queue"
	workflowName  = "scheduler_workflow"
	activityName  = "scheduler_activity"
)

type Config struct {
	Worker   temporalclient.WorkerConfig   `koanf:"worker"`
	Workflow temporalclient.WorkflowConfig `koanf:"workflow"`
	Activity temporalclient.ActivityConfig `koanf:"activity"`
}

var Configs = configset.Set[Config]{
	Default: &Config{},
}

type Scheduler struct {
	cfg        *Config
	temporal   temporalclient.Client
	sl         *zap.SugaredLogger
	dispatcher sdkservices.Dispatcher
	triggers   sdkservices.Triggers
}

func New(l *zap.Logger, tc temporalclient.Client, cfg *Config) *Scheduler {
	return &Scheduler{sl: l.Sugar(), temporal: tc, cfg: cfg}
}

func (sch *Scheduler) Start(ctx context.Context, dispatcher sdkservices.Dispatcher, triggers sdkservices.Triggers) error {
	sch.dispatcher = dispatcher
	sch.triggers = triggers

	w := temporalclient.NewWorker(sch.sl.Desugar(), sch.temporal.TemporalClient(), taskQueueName, sch.cfg.Worker)
	if w == nil {
		return nil
	}

	w.RegisterWorkflowWithOptions(sch.workflow, workflow.RegisterOptions{Name: workflowName})
	w.RegisterActivityWithOptions(sch.activity, activity.RegisterOptions{Name: activityName})

	if err := w.Start(); err != nil {
		return fmt.Errorf("schedule wf: worker start: %w", err)
	}

	return nil
}

func (sch *Scheduler) Create(ctx context.Context, tid sdktypes.TriggerID, schedule string) error {
	l := sch.sl.With("trigger_id", tid.String())

	_, err := sch.temporal.TemporalClient().ScheduleClient().Create(
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

	scheduleHandle := sch.temporal.TemporalClient().ScheduleClient().GetHandle(ctx, tid.String()) // validity of scheduleID is not checked by temporal
	if err := scheduleHandle.Delete(ctx); err != nil {
		return fmt.Errorf("schedule: delete scheduler workflow: %w", err)
	}

	sl.Infof("deleted schedule workflow for %s", tid)

	return nil
}

func (sch *Scheduler) Update(ctx context.Context, tid sdktypes.TriggerID, schedule string) error {
	sl := sch.sl.With("trigger_id", tid)

	h := sch.temporal.TemporalClient().ScheduleClient().GetHandle(ctx, tid.String()) // validity of scheduleID is not checked by temporal
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

func (sch *Scheduler) activity(ctx context.Context, tid sdktypes.TriggerID) error {
	sl := sch.sl.With("trigger_id", tid)

	ctx = authcontext.SetAuthnSystemUser(ctx)

	t, err := sch.triggers.Get(ctx, tid)
	if err != nil {
		if !errors.Is(err, sdkerrors.ErrNotFound) {
			return temporalclient.TranslateError(err, "get trigger %v", tid)
		}
	}

	if !t.IsValid() {
		sl.Warnf("trigger %v not found, removing schedule", tid)

		if err := sch.Delete(ctx, tid); err != nil {
			return temporalclient.TranslateError(err, "delete schedule for %v", tid)
		}

		return nil
	}

	eid, err := sch.dispatcher.Dispatch(ctx, sdktypes.NewEvent(tid).WithType("tick"), nil)
	if err != nil {
		return temporalclient.TranslateError(err, "dispatch event for %v", tid)
	}

	sl.With("event_id", eid).Infof("schedule event workflow for %v dispatched as %v", tid, eid)

	return nil
}

func (sch *Scheduler) workflow(wctx workflow.Context, tid sdktypes.TriggerID) error {
	return workflow.ExecuteActivity(
		temporalclient.WithActivityOptions(wctx, taskQueueName, sch.cfg.Activity),
		activityName,
		tid,
	).Get(wctx, nil)
}
