package dashboardsvc

import (
	"net/http"

	"github.com/gorilla/mux"

	"gitlab.com/softkitteh/autokitteh/internal/app/dashboardsvc/templates"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsrcsstore"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/eventsstore"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/projectsstore"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/secretsstore"
	"gitlab.com/softkitteh/autokitteh/internal/pkg/statestore"
	"gitlab.com/softkitteh/autokitteh/pkg/tmplrender"
)

type Config struct {
	TemplatesPath string `envconfig:"TEMPLATES_PATH" json:"TEMPLATES_PATH"`
}

type Svc struct {
	Config            Config
	EventsStore       eventsstore.Store
	ProjectsStore     projectsstore.Store
	EventSourcesStore eventsrcsstore.Store
	StateStore        statestore.Store
	SecretsStore      *secretsstore.Store

	renderFn tmplrender.RenderFunc
}

func (s *Svc) render(w http.ResponseWriter, name string, ctx interface{}) {
	if s.renderFn == nil {
		s.renderFn = tmplrender.New(s.Config.TemplatesPath, templates.FS).Render
	}

	s.renderFn(w, name, ctx)
}

func (s *Svc) static(w http.ResponseWriter, req *http.Request) {
	s.render(w, req.URL.Path, nil)
}

func (s *Svc) Register(r *mux.Router) {
	dashboard := r.PathPrefix("/dashboard/").Subrouter()
	dashboard.PathPrefix("/static/").Handler(http.StripPrefix("/dashboard/static/", http.HandlerFunc(s.static)))

	s.registerEvents(dashboard)
	s.registerEventSrcs(dashboard)
	s.registerProjects(dashboard)
}
