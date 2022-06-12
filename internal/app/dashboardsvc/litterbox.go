package dashboardsvc

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/autokitteh/autokitteh/examples/litterboxes"
	"github.com/autokitteh/autokitteh/internal/pkg/litterbox"
)

func (s *Svc) registerLitterbox(r *mux.Router) {
	lb := r.PathPrefix("/litterbox").Subrouter()
	lb.Path("").HandlerFunc(s.litterbox)
	lb.Path("/{id}").HandlerFunc(s.litterbox)
}

func (s *Svc) litterbox(w http.ResponseWriter, r *http.Request) {
	id := litterbox.LitterBoxID(mux.Vars(r)["id"])

	var src string

	if id != "" {
		bs, err := s.LitterBox.Get(r.Context(), id)
		if err != nil {
			if errors.Is(err, litterbox.ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			return
		}

		if bs != nil {
			src = string(bs)
		}
	}

	ctx := struct {
		Addr         string
		Examples     interface{}
		JSONExamples string
		ID           string
		Source       string
	}{
		Addr:         fmt.Sprintf("127.0.0.1:%d", s.Port),
		Examples:     litterboxes.Examples,
		JSONExamples: litterboxes.JSONExamples,
		ID:           string(id),
		Source:       string(src),
	}

	s.render(w, "litterbox.html", ctx)
}
