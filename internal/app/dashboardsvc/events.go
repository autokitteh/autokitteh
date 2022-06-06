package dashboardsvc

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	pbeventsrc "go.autokitteh.dev/idl/go/eventsrc"

	"go.autokitteh.dev/sdk/api/apievent"
	"go.autokitteh.dev/sdk/api/apilang"
	"go.autokitteh.dev/sdk/api/apiproject"
	"go.autokitteh.dev/sdk/api/apivalues"

	"github.com/autokitteh/autokitteh/internal/pkg/eventsstore"
)

func (s *Svc) registerEvents(r *mux.Router) {
	events := r.PathPrefix("/events").Subrouter()
	events.Path("/").HandlerFunc(s.listEvents)
	events.Path("").HandlerFunc(s.listEvents)
	events.Path("/{id:[^\\.]+}.json").HandlerFunc(s.rawEvent)
	events.Path("/{id}").HandlerFunc(s.event)
	events.Path("/{id}/projects/{pid}").HandlerFunc(s.eventForProject)
}

func (s *Svc) listEvents(w http.ResponseWriter, r *http.Request) {
	var (
		ofs, ln uint64
		pid     *apiproject.ProjectID
		err     error
	)

	Q := func(s string) string { return r.URL.Query().Get(s) }

	if s := Q("ofs"); s != "" {
		if ofs, err = strconv.ParseUint(s, 10, 32); err != nil {
			http.Error(w, "invalid ofs", http.StatusBadRequest)
			return
		}
	}

	if s := Q("len"); s != "" {
		if ln, err = strconv.ParseUint(s, 10, 32); err != nil {
			http.Error(w, "invalid len", http.StatusBadRequest)
			return
		}
	}

	if s := Q("pid"); s != "" {
		pid_ := apiproject.ProjectID(s)
		pid = &pid_
	}

	rs, err := s.EventsStore.List(r.Context(), pid, uint32(ofs), uint32(ln))
	if err != nil {
		http.Error(w, fmt.Sprintf("list error: %v", err), http.StatusInternalServerError)
		return
	}

	pbrs := make([]interface{}, len(rs))
	for i, r := range rs {
		pbr := struct {
			Event         interface{}
			UnwrappedData interface{}
			States        []interface{}
		}{
			Event:         r.Event.PB(),
			States:        make([]interface{}, len(r.States)),
			UnwrappedData: apivalues.UnwrapValuesMap(r.Event.Data(), apivalues.WithUnwrapJSONSafe()),
		}

		for j, state := range r.States {
			pbr.States[j] = state.PB()
		}

		pbrs[i] = pbr
	}

	var pbbindings map[string]interface{}
	if pid != nil {
		bindings, err := s.EventSourcesStore.GetProjectBindings(r.Context(), nil, pid, "", "", false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pbbindings = make(map[string]interface{}, len(bindings))
		for _, b := range bindings {
			pbbindings[b.EventSourceID().String()] = b.PB()
		}
	}

	ctx := struct {
		Records, Bindings interface{}
		Ofs, Len          uint64
		ProjectID         *apiproject.ProjectID
	}{
		Ofs:       ofs,
		Len:       ln,
		ProjectID: pid,
		Records:   pbrs,
		Bindings:  pbbindings,
	}

	s.render(w, "list-events.html", ctx)
}

func (s *Svc) rawEvent(w http.ResponseWriter, r *http.Request) {
	id := apievent.EventID(mux.Vars(r)["id"])

	ev, err := s.EventsStore.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, eventsstore.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "protodump.json.html", ev.PB())
}

func (s *Svc) event(w http.ResponseWriter, r *http.Request) {
	id := apievent.EventID(mux.Vars(r)["id"])

	ev, err := s.EventsStore.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, eventsstore.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	states, err := s.EventsStore.GetState(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("state: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	var ignoredpids, attnpids, pids []string

	pidsStrings := func(pids []apiproject.ProjectID) (l []string) {
		for _, pid := range pids {
			l = append(l, pid.String())
		}
		return
	}

	pbstates := make([]interface{}, len(states))
	for i, s := range states {
		pbstates[i] = s.PB()

		switch st := s.State().Get().(type) {
		case *apievent.ProcessingEventState:
			pids = pidsStrings(st.ProjectIDs())
			ignoredpids = pidsStrings(st.IgnoredProjectIDs())
		case *apievent.ProcessedEventState:
			pids = pidsStrings(st.ProjectIDs())
			attnpids = pidsStrings(st.AttnProjectIDs())
		}
	}

	assoc := ev.AssociationToken()
	srcID := ev.EventSourceID()

	// [# litterbox-assoc #]
	if srcID.EventSourceName() == "litterbox" {
		assoc, _, _ = strings.Cut(assoc, ",")
	}

	bindings, err := s.EventSourcesStore.GetProjectBindings(r.Context(), &srcID, nil, "", assoc, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pbbindings := make(map[string]interface{}, len(bindings))
	for _, b := range bindings {
		pbbindings[b.EventSourceID().String()] = b.PB()
	}

	s.render(w, "event.html", struct {
		Event, UnwrappedData, States, Bindings,
		ProjectIDs, IgnoredProjectIDs, AttnProjectIDs interface{}
	}{
		Event:             ev.PB(),
		UnwrappedData:     apivalues.UnwrapValuesMap(ev.Data(), apivalues.WithUnwrapJSONSafe()),
		States:            pbstates,
		ProjectIDs:        pids,
		IgnoredProjectIDs: ignoredpids,
		AttnProjectIDs:    attnpids,
		Bindings:          pbbindings,
	})
}

func (s *Svc) eventForProject(w http.ResponseWriter, r *http.Request) {
	id, pid := apievent.EventID(mux.Vars(r)["id"]), apiproject.ProjectID(mux.Vars(r)["pid"])

	proj, err := s.ProjectsStore.Get(r.Context(), pid)
	if err != nil {
		if errors.Is(err, eventsstore.ErrNotFound) {
			http.Error(w, "project not found", http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("projects: %v", err), http.StatusInternalServerError)
		return
	}

	ev, err := s.EventsStore.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, eventsstore.ErrNotFound) {
			http.Error(w, "event not found", http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("events: %v", err), http.StatusInternalServerError)
		return
	}

	states, err := s.EventsStore.GetStateForProject(r.Context(), id, pid)
	if err != nil {
		if errors.Is(err, eventsstore.ErrNotFound) {
			http.Error(w, "event for project not found", http.StatusNotFound)
			return
		}

		http.Error(w, fmt.Sprintf("project state: %v", err), http.StatusInternalServerError)
		return
	}

	srcID := ev.EventSourceID()

	assoc := ev.AssociationToken()
	var bindingName string

	// [# litterbox-assoc #]
	if srcID.EventSourceName() == "litterbox" {
		assoc, bindingName, _ = strings.Cut(assoc, ",")
	}

	bindings, err := s.EventSourcesStore.GetProjectBindings(r.Context(), &srcID, &pid, "", assoc, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var pbbinding *pbeventsrc.EventSourceProjectBinding
	if len(bindings) == 1 {
		pbbinding = bindings[0].PB()
	}

	if bindingName != "" && pbbinding != nil {
		pbbinding.Name = bindingName
	}

	var sum *apilang.RunSummary
	pbstates := make([]interface{}, len(states))
	for i, s := range states {
		pbstates[i] = s.PB()

		if sum == nil {
			switch sv := s.State().Get().(type) {
			case *apievent.ErrorProjectEventState:
				sum = sv.RunSummary()
			case *apievent.ProcessedProjectEventState:
				sum = sv.RunSummary()
			case *apievent.WaitingProjectEventState:
				sum = sv.RunSummary()
			}
		}
	}

	fsum := sum.Flatten()
	flog, fprints := fsum.PB().Log, fsum.PB().Prints

	s.render(w, "project-event.html", struct {
		Event, UnwrappedData, States, Binding, Project, RunSummary, FlatLog, FlatPrints interface{}
	}{
		Event:         ev.PB(),
		UnwrappedData: apivalues.UnwrapValuesMap(ev.Data(), apivalues.WithUnwrapJSONSafe()),
		States:        pbstates,
		Binding:       pbbinding,
		Project:       proj.PB(),
		RunSummary:    sum.PB(),
		FlatLog:       flog,
		FlatPrints:    fprints,
	})
}
