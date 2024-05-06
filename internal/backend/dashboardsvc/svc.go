package dashboardsvc

import (
	"net/http"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type Svc struct {
	fx.In

	Svcs  sdkservices.Services
	Muxes *muxes.Muxes
}

func Init(svc Svc) {
	svc.initIndex()
	svc.initObjects()
	svc.initProjects()
	svc.initConnections()
	svc.initIntegrations()
	svc.initEnvs()
	svc.initTriggers()
	svc.initDeployments()
	svc.initBuilds()
	svc.initEvents()
	svc.initSessions()
	svc.initAuth()
}

func (s *Svc) initIndex() {
	s.Muxes.NoAuth.HandleFunc("/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/projects", http.StatusFound)
	})
}
