package dashboardsvc

import (
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initConnections() {
	s.Muxes.Aux.Auth.HandleFunc("GET /connections", s.connections)
	s.Muxes.Aux.Auth.HandleFunc("GET /connections/{cid}", s.connection)
	s.Muxes.Aux.Auth.HandleFunc("DELETE /connections/{id}/vars", s.rmAllConnectionVars)
}

type connection struct{ sdktypes.Connection }

func (p connection) FieldsOrder() []string {
	return []string{"connection_id", "name", "project_id", "integration_id"}
}

func (p connection) HideFields() []string { return nil }

func (p connection) ExtraFields() map[string]any {
	var status string

	if s := p.Connection.Status(); s.IsValid() {
		text := s.Code().String()
		if s.Message() != "" {
			text += ": " + s.Message()
		}

		status = text
	}

	return map[string]any{"status": status}
}

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

	return genListData(f, kittehs.Transform(sdkCs, toConnection), drops...), nil
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
		status := http.StatusInternalServerError
		if err == sdkerrors.ErrNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
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
	if !sdkI.IsValid() {
		http.Error(w, "Integration not found", http.StatusNotFound)
		return
	}

	cvars, err := s.genVarsList(w, r, sdktypes.NewVarScopeID(cid))
	if err != nil {
		return
	}

	events, err := s.listEvents(w, r, sdkservices.ListEventsFilter{
		DestinationID: sdktypes.NewEventDestinationID(sdkC.ID()),
	})
	if err != nil {
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "connection.html", struct {
		MainURL string
		Title   string
		ID      string
		Name    string
		JSON    template.HTML
		Vars    list
		Events  list
		Caps    any
	}{
		MainURL: s.Muxes.MainURL,
		Title:   "Connection: " + p.Name().String(),
		ID:      cid.String(),
		Name:    p.Name().String(),
		JSON:    marshalObject(sdkC.ToProto()),
		Vars:    cvars,
		Events:  events,
		Caps:    p.Capabilities().ToProto(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
