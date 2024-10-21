package dashboardsvc

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initEvents() {
	s.Muxes.Aux.Auth.HandleFunc("/events", s.events)
	s.Muxes.Aux.Auth.HandleFunc("/events/{eid}", s.event)
	s.Muxes.Aux.Auth.HandleFunc("/events/{eid}/redispatch", s.redispatchEvent)
}

type event struct{ sdktypes.Event }

func (p event) FieldsOrder() []string {
	return []string{"created_at", "event_id", "connection_id", "integration_id"}
}

func (p event) HideFields() []string        { return nil }
func (p event) ExtraFields() map[string]any { return nil }

func toEvent(sdkP sdktypes.Event) event { return event{sdkP} }

func (s Svc) listEvents(w http.ResponseWriter, r *http.Request, f sdkservices.ListEventsFilter) (list, error) {
	if f.Limit <= 0 {
		f.Limit = getQueryNum(r, "events_limit", 50)
	}

	if f.MinSequenceNumber == 0 {
		f.MinSequenceNumber = uint64(getQueryNum(r, "events_min_seq", 0))
	}

	sdkCs, err := s.Svcs.Events().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	var drops []string
	if f.IntegrationID.IsValid() {
		drops = append(drops, "integration_id")
	}

	if f.DestinationID.IsValid() {
		drops = append(drops, "destination_id")
	}

	return genListData(f, kittehs.Transform(sdkCs, toEvent), drops...), nil
}

func (s Svc) events(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listEvents(w, r, sdkservices.ListEventsFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "events", ts)
}

func (s Svc) event(w http.ResponseWriter, r *http.Request) {
	eid, err := sdktypes.ParseEventID(r.PathValue("eid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkE, err := s.Svcs.Events().Get(r.Context(), eid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vw := sdktypes.DefaultValueWrapper
	vw.SafeForJSON = true
	vw.IgnoreFunctions = true

	data, err := vw.UnwrapMap(sdkE.Data())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	actives, err := s.Svcs.Deployments().List(r.Context(), sdkservices.ListDeploymentsFilter{
		State: sdktypes.DeploymentStateActive,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	testings, err := s.Svcs.Deployments().List(r.Context(), sdkservices.ListDeploymentsFilter{
		State: sdktypes.DeploymentStateTesting,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projects, err := s.Svcs.Projects().List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	envs, err := s.Svcs.Envs().List(r.Context(), sdktypes.InvalidProjectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// name -> deployment id
	deployments := kittehs.ListToMap(append(actives, testings...), func(d sdktypes.Deployment) (string, string) {
		var name string

		_, e := kittehs.FindFirst(envs, func(e sdktypes.Env) bool { return e.ID() == d.EnvID() })
		if e.IsValid() {
			name = "/" + e.Name().String()
			_, p := kittehs.FindFirst(projects, func(p sdktypes.Project) bool { return p.ID() == e.ProjectID() })
			if p.IsValid() {
				name = p.Name().String() + name
			}
		} else {
			name = d.ID().String()
		}

		return fmt.Sprintf("%s [%v]", name, d.State()), d.ID().String()
	})

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "event.html", struct {
		Title       string
		ID          string
		EventJSON   template.HTML
		DataJSON    template.HTML
		Deployments any
	}{
		Title:       "Event: " + sdkE.ID().String(),
		ID:          sdkE.ID().String(),
		EventJSON:   marshalObject(sdkE.WithData(nil).ToProto()),
		DataJSON:    template.HTML(jsonData),
		Deployments: deployments,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s Svc) redispatchEvent(w http.ResponseWriter, r *http.Request) {
	eid, err := sdktypes.ParseEventID(r.PathValue("eid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	did, err := sdktypes.ParseDeploymentID(r.URL.Query().Get("did"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	opts := sdkservices.DispatchOptions{
		Env:          r.URL.Query().Get("env"),
		DeploymentID: did,
	}

	eid1, err := s.Svcs.Dispatcher().Redispatch(r.Context(), eid, &opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/events/"+eid1.String(), http.StatusSeeOther)
}
