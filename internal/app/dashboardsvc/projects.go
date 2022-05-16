package dashboardsvc

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.com/softkitteh/autokitteh/pkg/autokitteh/api/apiproject"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/projectsstore"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/statestore"
)

func (s *Svc) registerProjects(r *mux.Router) {
	events := r.PathPrefix("/projects").Subrouter()
	events.Path("/{id}").HandlerFunc(s.project)
	events.Path("/{id}/state").HandlerFunc(s.projectStateList)
	events.Path("/{id}/secrets").HandlerFunc(s.projectSecretsList)
	events.Path("/{id}/state/{name}").HandlerFunc(s.projectState)
}

func (s *Svc) project(w http.ResponseWriter, r *http.Request) {
	id := apiproject.ProjectID(mux.Vars(r)["id"])

	proj, err := s.ProjectsStore.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, projectsstore.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "protodump.json.html", proj.PB())
}

func (s *Svc) projectSecretsList(w http.ResponseWriter, r *http.Request) {
	id := apiproject.ProjectID(mux.Vars(r)["id"])

	ns, err := s.SecretsStore.List(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "dump.json.html", ns)
}

func (s *Svc) projectStateList(w http.ResponseWriter, r *http.Request) {
	id := apiproject.ProjectID(mux.Vars(r)["id"])

	ns, err := s.StateStore.List(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "project-state.html", struct {
		ProjectID string
		Names     []string
	}{
		ProjectID: id.String(),
		Names:     ns,
	})
}

func (s *Svc) projectState(w http.ResponseWriter, r *http.Request) {
	id := apiproject.ProjectID(mux.Vars(r)["id"])
	name := mux.Vars(r)["name"]

	v, _, err := s.StateStore.Get(r.Context(), id, name)
	if err != nil {
		if errors.Is(err, statestore.ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.render(w, "dump.json.html", v)
}
