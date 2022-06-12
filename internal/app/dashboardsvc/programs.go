package dashboardsvc

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"go.autokitteh.dev/sdk/api/apiprogram"
	"go.autokitteh.dev/sdk/api/apiproject"
)

func (s *Svc) registerPrograms(r *mux.Router) {
	events := r.PathPrefix("/programs").Subrouter()
	events.Path("/{pid}").HandlerFunc(s.program)
}

func (s *Svc) program(w http.ResponseWriter, r *http.Request) {
	pid := apiproject.ProjectID(mux.Vars(r)["pid"])

	rawPath := r.URL.Query().Get("path")

	path, err := apiprogram.ParsePathString(rawPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	f, err := s.Programs.Get(r.Context(), pid, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if f == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	fmt.Fprintf(w, "# FetchedAt: %v\n\n", f.FetchedAt)

	w.Write(f.Source)
}
