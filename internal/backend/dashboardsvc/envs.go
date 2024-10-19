package dashboardsvc

import (
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initEnvs() {
	s.Muxes.Aux.Auth.HandleFunc("/envs", s.envs)
	s.Muxes.Aux.Auth.HandleFunc("/envs/{eid}", s.env)
}

type env struct{ sdktypes.Env }

func (p env) FieldsOrder() []string       { return []string{"env_id", "name"} }
func (p env) HideFields() []string        { return nil }
func (p env) ExtraFields() map[string]any { return nil }

func toEnv(sdkE sdktypes.Env) env { return env{sdkE} }

func (s Svc) listEnvs(w http.ResponseWriter, r *http.Request, pid sdktypes.ProjectID) (list, error) {
	sdkEs, err := s.Svcs.Envs().List(r.Context(), pid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	var drops []string
	if pid.IsValid() {
		drops = append(drops, "project_id")
	}

	return genListData(pid.String(), kittehs.Transform(sdkEs, toEnv), drops...), nil
}

func (s Svc) envs(w http.ResponseWriter, r *http.Request) {
	l, err := s.listEnvs(w, r, sdktypes.InvalidProjectID)
	if err != nil {
		return
	}

	renderList(w, r, "envs", l)
}

func (s Svc) env(w http.ResponseWriter, r *http.Request) {
	eid, err := sdktypes.StrictParseEnvID(r.PathValue("eid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkE, err := s.Svcs.Envs().GetByID(r.Context(), eid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	evars, err := s.genVarsList(w, r, sdktypes.NewVarScopeID(eid))
	if err != nil {
		return
	}

	trgs, err := s.listTriggers(w, r, sdkservices.ListTriggersFilter{
		EnvID: eid,
	})
	if err != nil {
		return
	}

	deps, err := s.listDeployments(w, r, sdkservices.ListDeploymentsFilter{EnvID: eid})
	if err != nil {
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "env.html", struct {
		Message     string
		Title       string
		Name        string
		JSON        template.HTML
		Vars        list
		Triggers    list
		Deployments list
	}{
		Title:       "Env: " + sdkE.Name().String(),
		Name:        sdkE.Name().String(),
		JSON:        marshalObject(sdkE.ToProto()),
		Vars:        evars,
		Triggers:    trgs,
		Deployments: deps,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
