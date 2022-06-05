package dashboardsvc

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/autokitteh/autokitteh/examples/litterboxes"
)

func (s *Svc) registerLitterbox(r *mux.Router) {
	lb := r.PathPrefix("/litterbox").Subrouter()
	lb.Path("").HandlerFunc(s.litterbox)
}

func (s *Svc) litterbox(w http.ResponseWriter, r *http.Request) {
	ctx := struct {
		Addr         string
		Examples     interface{}
		JSONExamples string
	}{
		Addr:         fmt.Sprintf("127.0.0.1:%d", s.Port),
		Examples:     litterboxes.Examples,
		JSONExamples: litterboxes.JSONExamples,
	}

	s.render(w, "litterbox.html", ctx)
}
