package dashboardsvc

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
)

type svc struct {
	sdkservices.Services
	*http.ServeMux
}

const rootPath = "/internal/dashboard/"

func Init(svcs sdkservices.Services, muxes *muxes.Muxes) {
	svc := &svc{Services: svcs, ServeMux: muxes.Auth}

	svc.HandleFunc(rootPath, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, rootPath+"projects", http.StatusTemporaryRedirect)
	})

	svc.initAuth()
	svc.initBuilds()
	svc.initConnections()
	svc.initDeployments()
	svc.initEnvs()
	svc.initEvents()
	svc.initIntegrations()
	svc.initObjects()
	svc.initProjects()
	svc.initSessions()
	svc.initToken()
	svc.initTriggers()
	svc.initVars()
}
