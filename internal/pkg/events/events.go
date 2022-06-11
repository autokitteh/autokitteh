package events

import (
	"context"
	"fmt"
	"time"

	temporalclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/proto"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apieventsrc"
	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/L"
	"github.com/autokitteh/pubsub"

	"github.com/autokitteh/autokitteh/internal/pkg/accountsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsrcsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
	"github.com/autokitteh/autokitteh/internal/pkg/projectsstore"
)

type Stores struct {
	Events       eventsstore.Store
	EventSources eventsrcsstore.Store
	Projects     projectsstore.Store
	Accounts     accountsstore.Store
}

type Events struct {
	Temporal temporalclient.Client
	PubSub   pubsub.PubSub
	Stores   Stores
	Run      func(workflow.Context, *apievent.Event, *apiproject.Project, string) (*apilang.RunSummary, error)

	worker worker.Worker

	L L.Nullable
}

func (e *Events) Init() {
	e.worker = worker.New(e.Temporal, "ingest-event", worker.Options{})
	e.worker.RegisterWorkflow(e.ingestEventWorkflow)
	e.worker.RegisterWorkflow(e.ingestProjectEventWorkflow)
}

func (e *Events) Start() error { return e.worker.Start() }

func GetIngestProjectEventWorkflowID(eid apievent.EventID, pid apiproject.ProjectID) string {
	return fmt.Sprintf("ingest_project_event-%v-%v", eid, pid)
}

func TopicForEvent(id apievent.EventID) string {
	return fmt.Sprintf("event-tracking-%s", id.String())
}

func TopicForProject(id apiproject.ProjectID) string {
	return fmt.Sprintf("project-tracking-%s", id.String())
}

func (e *Events) MonitorProjectEvents(
	ctx context.Context,
	ch chan<- *apievent.TrackIngestEventUpdate,
	pid apiproject.ProjectID,
) error {
	l := e.L.With("project_id", pid.String())

	sub, err := e.PubSub.Subscribe(ctx, TopicForProject(pid))
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	defer func() {
		go func() {
			if err := sub.Unsubscribe(context.Background()); err != nil {
				e.L.Error("unsubscribe error", "err", err)
			}
		}()
	}()

	for {
		payload, err := sub.Consume(ctx)
		if err != nil {
			close(ch)

			switch err {
			case pubsub.ErrUnsubscribed:
				l.Debug("unsubscribed")
			case context.Canceled:
				l.Debug("canceled")
			default:
				return L.Error(l, "consume error", "err", err)
			}

			return nil
		}

		var pb apievent.TrackIngestEventUpdatePB

		if err := proto.Unmarshal(payload, &pb); err != nil {
			l.Error("update unmarshal error", "err", err)
			continue
		}

		upd, err := apievent.TrackIngestEventUpdateFromProto(&pb)
		if err != nil {
			l.Error("invalid update", "err", err)
			continue
		}

		l.Debug("sent update", "upd", upd)

		ch <- upd
	}
}

func (e *Events) TrackIngestEvent(
	ctx context.Context,
	ch chan<- *apievent.TrackIngestEventUpdate,
	id apievent.EventID,
	srcid apieventsrc.EventSourceID,
	assoc string,
	originalID string,
	typ string,
	data map[string]*apivalues.Value,
	memo map[string]string,
) error {
	if id == "" {
		id = apievent.NewEventID()
	}

	l := e.L.With("event_id", id.String())

	sub, err := e.PubSub.Subscribe(ctx, TopicForEvent(id))
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	defer func() {
		go func() {
			if err := sub.Unsubscribe(context.Background()); err != nil {
				e.L.Error("unsubscribe error", "err", err)
			}
		}()
	}()

	we, err := e.ingestEvent(ctx, id, srcid, assoc, originalID, typ, data, memo)
	if err != nil {
		return err
	}

	go func() {
		for {
			payload, err := sub.Consume(ctx)
			if err != nil {
				close(ch)

				switch err {
				case pubsub.ErrUnsubscribed:
					l.Debug("unsubscribed")
				case context.Canceled:
					l.Debug("canceled")
				default:
					l.Error("consume error", "err", err)
				}

				return
			}

			var pb apievent.TrackIngestEventUpdatePB

			if err := proto.Unmarshal(payload, &pb); err != nil {
				l.Error("update unmarshal error", "err", err)
				continue
			}

			upd, err := apievent.TrackIngestEventUpdateFromProto(&pb)
			if err != nil {
				l.Error("invalid update", "err", err)
				continue
			}

			l.Debug("sent update", "upd", upd)

			ch <- upd
		}
	}()

	if err := we.Get(ctx, nil); err != nil {
		return L.Error(l, "wait error", "err", err)
	}

	return nil
}

func (e *Events) IngestEvent(
	ctx context.Context,
	id apievent.EventID,
	srcid apieventsrc.EventSourceID,
	assoc string,
	originalID string,
	typ string,
	data map[string]*apivalues.Value,
	memo map[string]string,
) (apievent.EventID, error) {
	if id == "" {
		id = apievent.NewEventID()
	}

	_, err := e.ingestEvent(ctx, id, srcid, assoc, originalID, typ, data, memo)
	return id, err
}

func (e *Events) ingestEvent(
	ctx context.Context,
	id apievent.EventID,
	srcid apieventsrc.EventSourceID,
	assoc string,
	originalID string,
	typ string,
	data map[string]*apivalues.Value,
	memo map[string]string,
) (temporalclient.WorkflowRun, error) {
	l := e.L.With("srcid", srcid, "type", typ, "original_id", originalID, "assoc", assoc)

	id, err := e.Stores.Events.Add(ctx, id, srcid, assoc, originalID, typ, data, memo)
	if err != nil {
		return nil, fmt.Errorf("add: %w", err)
	}

	l.Debug("event added", "id", id)

	if err := e.Stores.Events.UpdateState(ctx, id, apievent.NewPendingEventState()); err != nil {
		return nil, L.Error(l, "update event state failed", "err", err)
	}

	wopts := temporalclient.StartWorkflowOptions{
		ID:        "ingest-event-" + id.String(),
		TaskQueue: "ingest-event",
	}

	we, err := e.Temporal.ExecuteWorkflow(ctx, wopts, e.ingestEventWorkflow, id, srcid, assoc)
	if err != nil {
		l.Error("start workflow error", "err", err)

		err = fmt.Errorf("start workflow: %w", err)

		go func() {
			if err := e.Stores.Events.UpdateState(ctx, id, apievent.NewErrorEventState(err)); err != nil {
				l.Error("update event state failed", "err", err)
			}
		}()

		return nil, err
	}

	l.Debug("started ingest-event workflow", "workflow_id", we.GetID(), "run_id", we.GetRunID())

	return we, nil
}

func (e *Events) ingestEventWorkflow(
	ctx workflow.Context,
	id apievent.EventID,
	srcid apieventsrc.EventSourceID,
	assoc string,
) error {
	l := e.L.With("event_id", id)

	ctx = workflow.WithLocalActivityOptions(
		ctx,
		workflow.LocalActivityOptions{
			ScheduleToCloseTimeout: 10 * time.Second,
		},
	)

	erd, err := e.getEventRelatedData(ctx, id, srcid, assoc)
	if err != nil {
		l.Error("getEventRelatedData failed", "err", err)
		err = fmt.Errorf("get event related data: %w", err)
		e.updateErrorState(ctx, id, err)
		return err
	}

	if r := erd.IgnoreReason; r != "" {
		l.Debug("ignored", "reason", r)
		e.updateState(ctx, id, apievent.NewIgnoredEventState(r))
		return nil
	}

	l.Debug("got event related data", "bindings", erd.Bindings)

	pids := struct {
		active              map[apiproject.ProjectID]string
		activePids, ignored []apiproject.ProjectID
	}{
		active: map[apiproject.ProjectID]string{},
	}

	for _, b := range erd.Bindings {
		pid := b.ProjectID()

		l := l.With("project_id", pid)

		if !b.Settings().Enabled() {
			l.Debug("ignored: binding disabled")
			pids.ignored = append(pids.ignored, pid)
			continue
		}

		p, ok := erd.Projects[pid]
		if !ok || !p.Settings().Enabled() {
			l.Debug("ignored: project disabled")
			pids.ignored = append(pids.ignored, pid)
			continue
		}

		a, ok := erd.Accounts[p.AccountName()]
		if !ok || !a.Settings().Enabled() {
			l.Debug("ignored: project account disabled")
			pids.ignored = append(pids.ignored, pid)
			continue
		}

		l.Debug("going to process")

		pids.active[pid] = b.Name()
		pids.activePids = append(pids.activePids, pid)
	}

	e.updateState(ctx, id, apievent.NewProcessingEventState(pids.activePids, pids.ignored))

	futs := make(map[apiproject.ProjectID]workflow.ChildWorkflowFuture, len(pids.active))

	for _, pid := range pids.ignored {
		e.updateProjectState(ctx, id, pid, apievent.NewIgnoredProjectEventState("disabled"))
	}

	for pid, bname := range pids.active {
		e.updateProjectState(ctx, id, pid, apievent.NewPendingProjectEventState())

		cwo := workflow.ChildWorkflowOptions{
			WorkflowID: GetIngestProjectEventWorkflowID(erd.Event.ID(), pid),
		}
		ctx = workflow.WithChildOptions(ctx, cwo)

		f := workflow.ExecuteChildWorkflow(ctx, e.ingestProjectEventWorkflow, erd.Event, erd.Projects[pid], bname)

		futs[pid] = f
	}

	var all, fails []apiproject.ProjectID

	// TODO: "parallelize" this, this waits serially, which should work, but it'll be
	//       nicer to have this using select or waitall.
	for pid, fut := range futs {
		l := l.With("pid", pid)
		all = append(all, pid)

		var state *apievent.ProjectEventState

		if err := fut.Get(ctx, &state); err != nil {
			l.Error("project processing error", "err", err)
			fails = append(fails, pid)
		} else {
			l.Debug("project processing done")

			if state.IsError() {
				l.Debug("project event in error state")
				fails = append(fails, pid)
			}
		}
	}

	l.Debug("event processing done")

	e.updateState(ctx, id, apievent.NewProcessedEventState(all, fails))

	return nil
}

func (e *Events) ingestProjectEventWorkflow(
	ctx workflow.Context,
	event *apievent.Event,
	project *apiproject.Project,
	bindingName string,
) (*apievent.ProjectEventState, error) {
	l := e.L.With("event_id", event.ID(), "project_id", project.ID())

	l.Debug("processing project event")

	ctx = workflow.WithLocalActivityOptions(
		ctx,
		workflow.LocalActivityOptions{
			ScheduleToCloseTimeout: 10 * time.Second,
		},
	)

	sum, err := e.Run(ctx, event, project, bindingName)
	if err != nil {
		l.Debug("run error", "err", err)
		state := apievent.NewErrorProjectEventState(err, sum)
		e.updateProjectState(ctx, event.ID(), project.ID(), state)
		return state, nil
	}

	l.Debug("session run completed", "summary", sum)

	state := apievent.NewCompletedProjectEventState(sum)

	e.updateProjectState(ctx, event.ID(), project.ID(), state)

	return state, nil
}
