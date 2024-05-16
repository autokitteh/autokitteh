package dashboardsvc

import (
	"fmt"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initConnections() {
	s.Muxes.Auth.HandleFunc("GET /connections", s.connections)
	s.Muxes.Auth.HandleFunc("GET /connections/{cid}", s.connection)
	s.Muxes.Auth.HandleFunc("GET /connections/{id}/init", s.initConnection)
	s.Muxes.Auth.HandleFunc("DELETE /connections/{id}/vars", s.rmAllConnectionVars)
}

type connection struct{ sdktypes.Connection }

func (p connection) FieldsOrder() []string {
	return []string{"connection_id", "name", "project_id", "integration_id"}
}

func (p connection) HideFields() []string { return nil }

func toConnection(sdkC sdktypes.Connection) connection { return connection{sdkC} }

func (s Svc) listConnections(w http.ResponseWriter, r *http.Request, f sdkservices.ListConnectionsFilter) (list, error) {
	sdkCs, err := s.Svcs.Connections().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	var drops []string
	if f.ProjectID.IsValid() {
		drops = append(drops, "project_id")
	}

	return genListData(kittehs.Transform(sdkCs, toConnection), drops...), nil
}

func (s Svc) connections(w http.ResponseWriter, r *http.Request) {
	l, err := s.listConnections(w, r, sdkservices.ListConnectionsFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "connections", l)
}

func (s Svc) getConnection(w http.ResponseWriter, r *http.Request) (sdktypes.Connection, bool) {
	cid, err := sdktypes.StrictParseConnectionID(r.PathValue("cid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return sdktypes.InvalidConnection, false
	}

	sdkC, err := s.Svcs.Connections().Get(r.Context(), cid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return sdktypes.InvalidConnection, false
	}

	if !sdkC.IsValid() {
		http.Error(w, "Connection not found", http.StatusNotFound)
		return sdktypes.InvalidConnection, false
	}

	return sdkC, true
}

func (s Svc) connection(w http.ResponseWriter, r *http.Request) {
	sdkC, ok := s.getConnection(w, r)
	if !ok {
		return
	}

	p := toConnection(sdkC)
	cid := sdkC.ID()

	sdkI, err := s.Svcs.Integrations().GetByID(r.Context(), sdkC.IntegrationID())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !sdkI.Get().IsValid() {
		http.Error(w, "Integration not found", http.StatusNotFound)
		return
	}

	cvars, err := s.genVarsList(w, r, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return
	}

	events, err := s.listEvents(w, r, sdkservices.ListEventsFilter{
		ConnectionID: sdkC.ID(),
	})
	if err != nil {
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "connection.html", struct {
		Title         string
		ID            string
		Name          string
		JSON          template.HTML
		Vars          list
		Events        list
		ConnectionURL string
	}{
		Title:         "Connection: " + p.Name().String(),
		ID:            cid.String(),
		Name:          p.Name().String(),
		JSON:          marshalObject(sdkC.ToProto()),
		Vars:          cvars,
		Events:        events,
		ConnectionURL: sdkI.Get().ConnectionURL().String(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s Svc) initUnknownConnection(w http.ResponseWriter, r *http.Request, id, vars string) {
	iid, err := sdktypes.StrictParseIntegrationID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	l, err := s.listConnections(w, r, sdkservices.ListConnectionsFilter{
		IntegrationID: iid,
	})
	if err != nil {
		return
	}

	l.Headers = append(l.Headers, "")

	for i := range l.Items {
		l.Items[i] = append(
			l.Items[i],
			template.HTML(
				fmt.Sprintf(`<button onclick="window.location.href='/connections/%s/init?vars=%s'">Select</button>`,
					l.UnformattedItems[i][0],
					vars,
				),
			),
		)
	}

	renderList(w, r, "connections", l)
}

func (s Svc) initConnection(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query().Get("vars")

	id := r.PathValue("id")
	if sdktypes.IsIntegrationID(id) {
		s.initUnknownConnection(w, r, id, vars)
		return
	}

	var data []sdktypes.Var
	if err := kittehs.DecodeURLData(vars, &data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cid, err := sdktypes.StrictParseConnectionID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data = kittehs.Transform(data, func(v sdktypes.Var) sdktypes.Var {
		return v.WithScopeID(sdktypes.NewVarScopeID(cid))
	})

	if err := s.Svcs.Vars().Set(r.Context(), data...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/connections/"+cid.String()+`?msg=Connection initialized 😸`, http.StatusSeeOther)
}

func (s Svc) rmAllConnectionVars(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	cid, err := sdktypes.StrictParseConnectionID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.Svcs.Vars().Delete(r.Context(), sdktypes.NewVarScopeID(cid)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
