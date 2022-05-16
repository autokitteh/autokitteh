package dashboardsvc

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apieventsrc"
	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsrcsstore"
)

func (s *Svc) registerEventSrcs(r *mux.Router) {
	events := r.PathPrefix("/eventsrcs").Subrouter()
	events.Path("/{id}").HandlerFunc(s.eventSrc)
	events.Path("/{id}/bindings/{pid}").HandlerFunc(s.binding)
	events.Path("/{id}/bindings/{pid}/").HandlerFunc(s.binding)
	events.Path("/{id}/bindings/{pid}/{name}").HandlerFunc(s.bindingName)
	events.Path("/{id}/bindings").HandlerFunc(s.bindings)
	events.Path("/{id}/bindings/").HandlerFunc(s.bindings)
}

func (s *Svc) eventSrc(w http.ResponseWriter, r *http.Request) {
	id := apieventsrc.EventSourceID(mux.Vars(r)["id"])

	ev, err := s.EventSourcesStore.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, eventsrcsstore.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "protodump.json.html", ev.PB())
}

func (s *Svc) binding(w http.ResponseWriter, r *http.Request) {
	var id *apieventsrc.EventSourceID
	if x := apieventsrc.EventSourceID(mux.Vars(r)["id"]); !x.Empty() {
		id = &x
	}

	pid := apiproject.ProjectID(mux.Vars(r)["pid"])

	bs, err := s.EventSourcesStore.GetProjectBindings(r.Context(), id, &pid, "", "", true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "dump.json.html", bs)
}

func (s *Svc) bindingName(w http.ResponseWriter, r *http.Request) {
	var id *apieventsrc.EventSourceID
	if x := apieventsrc.EventSourceID(mux.Vars(r)["id"]); !x.Empty() {
		id = &x
	}

	pid := apiproject.ProjectID(mux.Vars(r)["pid"])
	name := mux.Vars(r)["name"]

	bs, err := s.EventSourcesStore.GetProjectBindings(r.Context(), id, &pid, name, "", true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(bs) == 0 {
		http.Error(w, "not found", http.StatusNotFound)
	}

	s.render(w, "protodump.json.html", bs[0].PB())
}

func (s *Svc) bindings(w http.ResponseWriter, r *http.Request) {
	var id *apieventsrc.EventSourceID
	if x := apieventsrc.EventSourceID(mux.Vars(r)["id"]); !x.Empty() {
		id = &x
	}

	bs, err := s.EventSourcesStore.GetProjectBindings(r.Context(), id, nil, "", "", true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "dump.json.html", bs)
}
