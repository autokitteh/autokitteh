package dashboardsvc

import (
	_ "embed"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initProjects() {
	s.Muxes.Auth.HandleFunc("/projects", s.projects)
	s.Muxes.Auth.HandleFunc("/projects/{pid}", s.project)
}

type project struct{ sdktypes.Project }

func (p project) FieldsOrder() []string       { return []string{"name", "project_id"} }
func (p project) HideFields() []string        { return nil }
func (p project) ExtraFields() map[string]any { return nil }

func toProject(sdkP sdktypes.Project) project { return project{sdkP} }

func (s Svc) listProjects(w http.ResponseWriter, r *http.Request) (list, error) {
	sdkPs, err := s.Svcs.Projects().List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(kittehs.Transform(sdkPs, toProject)), nil
}

func (s Svc) projects(w http.ResponseWriter, r *http.Request) {
	ps, err := s.listProjects(w, r)
	if err != nil {
		return
	}

	renderList(w, r, "projects", ps)
}

func (s Svc) project(w http.ResponseWriter, r *http.Request) {
	pid, err := sdktypes.StrictParseProjectID(r.PathValue("pid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkP, err := s.Svcs.Projects().GetByID(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := toProject(sdkP)

	cs, err := s.listConnections(w, r, sdkservices.ListConnectionsFilter{
		ProjectID: pid,
	})
	if err != nil {
		return
	}

	es, err := s.listEnvs(w, r, pid)
	if err != nil {
		return
	}

	ts, err := s.listTriggers(w, r, sdkservices.ListTriggersFilter{ProjectID: pid})
	if err != nil {
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "project.html", struct {
		Message     string
		Title       string
		Name        string
		JSON        template.HTML
		Connections list
		Envs        list
		Triggers    list
		Sessions    list
	}{
		Title:       "Project: " + p.Name().String(),
		Name:        p.Name().String(),
		JSON:        marshalObject(sdkP.ToProto()),
		Connections: cs,
		Envs:        es,
		Triggers:    ts,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
