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
	svc.initAuth()
	svc.initBuilds()
	svc.initConnections()
	svc.initDeployments()
	svc.initEnvs()
	svc.initEvents()
	svc.initIndex()
	svc.initIntegrations()
	svc.initObjects()
	svc.initProjects()
	svc.initSessions()
	svc.initToken()
	svc.initTriggers()
	svc.initVars()
}

func (s *Svc) initIndex() {
	s.Muxes.NoAuth.HandleFunc("/dashboard/{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard/projects", http.StatusFound)
	})
}
