package connectionsinitsvc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/fx"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

type Svc struct{ svcs Svcs }

type Svcs struct {
	fx.In

	Muxes        *muxes.Muxes
	Connections  sdkservices.Connections
	Integrations sdkservices.Integrations
	Vars         sdkservices.Vars
}

func New(svcs Svcs) Svc { return Svc{svcs: svcs} }

func Init(svcs Svcs) {
	s := Svc{svcs}

	// These paths should correspond to the ones enriched in the connection service.
	svcs.Muxes.Auth.HandleFunc("POST /connections/{id}/test", s.testConnection)
	svcs.Muxes.Auth.HandleFunc("POST /connections/{id}/refresh", s.refreshConnection)
	svcs.Muxes.Auth.HandleFunc("GET /connections/{id}/init/{origin}", s.init)
	svcs.Muxes.Auth.HandleFunc("GET /connections/{id}/init", s.init)
	svcs.Muxes.Auth.HandleFunc("GET /connections/{id}/postinit", s.postInit)
	svcs.Muxes.Auth.HandleFunc("GET /connections/{id}/success", s.success)
}

func (s Svc) init(w http.ResponseWriter, r *http.Request) {
	id, origin := r.PathValue("id"), r.PathValue("origin")

	cid, err := sdktypes.StrictParseConnectionID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := s.svcs.Connections.Get(r.Context(), cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !conn.IsValid() {
		http.Error(w, "connection not found", http.StatusNotFound)
		return
	}

	integ, err := s.svcs.Integrations.GetByID(r.Context(), conn.IntegrationID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !integ.IsValid() {
		http.Error(w, "integration not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("%s?cid=%v&origin=%s", integ.ConnectionURL(), cid, origin), http.StatusFound)
}

// postInit is the last step in the connection initialization flow.
// It saves the connection's data in the connection's scope, and
// redirects the user to the last HTTP response, based on its origin
// (local AK server / local VS Code extension / SaaS web UI).
func (s Svc) postInit(w http.ResponseWriter, r *http.Request) {
	vars, origin := r.URL.Query().Get("vars"), r.URL.Query().Get("origin")

	var data []sdktypes.Var
	if err := kittehs.DecodeURLData(vars, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cid, err := sdktypes.StrictParseConnectionID(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data = kittehs.Transform(data, func(v sdktypes.Var) sdktypes.Var {
		return v.WithScopeID(sdktypes.NewVarScopeID(cid))
	})

	if err := s.svcs.Vars.Set(r.Context(), data...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var u string
	switch origin {
	case "vscode":
		u = "vscode://autokitteh.autokitteh?cid=%s"
	default:
		// Another redirect just to get rid of the secrets in the URL.
		u = "/connections/%s/success"
	}
	http.Redirect(w, r, fmt.Sprintf(u, cid), http.StatusFound)
}

func (s Svc) testConnection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cid, err := sdktypes.StrictParseConnectionID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	st, err := s.svcs.Connections.Test(r.Context(), cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(st); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s Svc) refreshConnection(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cid, err := sdktypes.StrictParseConnectionID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	st, err := s.svcs.Connections.RefreshStatus(r.Context(), cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(st); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/connections/%s", id), http.StatusFound)
}

func (s Svc) success(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "success - you can now close this tab.")
}
