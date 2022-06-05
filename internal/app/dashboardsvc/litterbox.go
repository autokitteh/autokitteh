package dashboardsvc

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Svc) registerLitterbox(r *mux.Router) {
	lb := r.PathPrefix("/litterbox").Subrouter()
	lb.Path("").HandlerFunc(s.litterbox)
}

func (s *Svc) litterbox(w http.ResponseWriter, r *http.Request) {
	ctx := struct {
		Addr string
	}{
		Addr: fmt.Sprintf("127.0.0.1:%d", s.Port),
	}

	s.render(w, "litterbox.html", ctx)
}
