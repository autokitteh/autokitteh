package events

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiaccount"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apievent"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"github.com/autokitteh/autokitteh/pkg/autokitteh/api/apiproject"
)

type eventRelatedData struct {
	Event        *apievent.Event
	SrcAccount   *apiaccount.Account
	Src          *apieventsrc.EventSource
	IgnoreReason string

	Bindings map[apiproject.ProjectID]*apieventsrc.EventSourceProjectBinding
	Projects map[apiproject.ProjectID]*apiproject.Project
	Accounts map[apiaccount.AccountName]*apiaccount.Account
}

func (e *Events) getEventRelatedData(ctx workflow.Context, id apievent.EventID, srcid apieventsrc.EventSourceID, assoc string) (*eventRelatedData, error) {
	var erd *eventRelatedData

	fut := workflow.ExecuteLocalActivity(
		ctx,
		func(
			ctx context.Context,
			id apievent.EventID,
			srcid apieventsrc.EventSourceID,
			assoc string,
		) (erd *eventRelatedData, err error) {
			stores := &e.Stores
			erd = &eventRelatedData{}

			l := e.L.With("event_id", id, "src_id", srcid)

			if erd.Src, err = stores.EventSources.Get(ctx, srcid); err != nil {
				err = fmt.Errorf("source %q: %w", srcid, err)
				return
			}

			if !erd.Src.Settings().Enabled() {
				erd.IgnoreReason = "event source disabled"
				return
			}

			if erd.SrcAccount, err = stores.Accounts.Get(ctx, erd.Src.ID().AccountName()); err != nil {
				err = fmt.Errorf("account %q: %w", erd.Src.ID().AccountName(), err)
				return
			}

			if !erd.SrcAccount.Settings().Enabled() {
				erd.IgnoreReason = "event source account disabled"
				return
			}

			if erd.Event, err = stores.Events.Get(ctx, id); err != nil {
				err = fmt.Errorf("event %q: %w", id, err)
				return
			}

			var bindings []*apieventsrc.EventSourceProjectBinding
			if bindings, err = stores.EventSources.GetProjectBindings(ctx, &srcid, nil, "", assoc, true); err != nil {
				err = fmt.Errorf("bindings %q: %w", srcid, err)
				return
			}

			erd.Bindings = make(map[apiproject.ProjectID]*apieventsrc.EventSourceProjectBinding)
			for _, b := range bindings {
				erd.Bindings[b.ProjectID()] = b
			}

			pids := make([]apiproject.ProjectID, 0, len(erd.Bindings))
			for pid, b := range erd.Bindings {
				l := l.With("project_id", pid)

				if !b.Settings().Enabled() {
					l.Debug("binding disabled")
					continue
				}

				// be extra sure ¯\_(ツ)_/¯
				if b.AssociationToken() != assoc {
					l.Debug("association token mismatch")
					continue
				}

				pids = append(pids, pid)
			}

			if erd.Projects, err = stores.Projects.BatchGet(ctx, pids); err != nil {
				err = fmt.Errorf("projects: %w", err)
				return
			}

			aids := make([]apiaccount.AccountName, 0, len(erd.Projects))
			for _, p := range erd.Projects {
				if !p.Settings().Enabled() {
					continue
				}

				aids = append(aids, p.AccountName())
			}

			accounts, err := stores.Accounts.BatchGet(ctx, aids)
			if err != nil {
				err = fmt.Errorf("accounts: %w", err)
				return
			}

			erd.Accounts = make(map[apiaccount.AccountName]*apiaccount.Account, len(accounts))
			for _, a := range accounts {
				erd.Accounts[a.Name()] = a
			}

			return
		},
		id, srcid, assoc,
	)

	if err := fut.Get(ctx, &erd); err != nil {
		return nil, fmt.Errorf("get: %w", err)
	}

	return erd, nil
}

func (e *Events) updateErrorState(ctx workflow.Context, id apievent.EventID, err error) {
	e.updateState(ctx, id, apievent.NewErrorEventState(err))
}

func (e *Events) updateState(ctx workflow.Context, id apievent.EventID, state *apievent.EventState) {
	l := e.L.With("event_id", id, "state", state)

	l.Debug("updating event state")

	if err := workflow.ExecuteLocalActivity(ctx, e.Stores.Events.UpdateState, id, state).Get(ctx, nil); err != nil {
		l.Error("update event state failed", "err", err)
	}
}

func (e *Events) updateProjectState(ctx workflow.Context, id apievent.EventID, pid apiproject.ProjectID, state *apievent.ProjectEventState) {
	l := e.L.With("event_id", id, "project_id", pid, "state", state)

	l.Debug("updating project state")

	if err := workflow.ExecuteLocalActivity(ctx, e.Stores.Events.UpdateStateForProject, id, pid, state).Get(ctx, nil); err != nil {
		l.Error("update project event state failed", "err", err)
	}
}
