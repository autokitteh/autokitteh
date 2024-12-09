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

func (s *svc) initEvents() {
	s.HandleFunc(rootPath+"events", s.events)
	s.HandleFunc(rootPath+"events/{eid}", s.event)
	s.HandleFunc(rootPath+"events/{eid}/redispatch", s.redispatchEvent)
}

type event struct{ sdktypes.Event }

func (p event) FieldsOrder() []string {
	return []string{"created_at", "event_id", "connection_id", "integration_id"}
}

func (p event) HideFields() []string        { return nil }
func (p event) ExtraFields() map[string]any { return nil }

func toEvent(sdkP sdktypes.Event) event { return event{sdkP} }

func (s *svc) listEvents(w http.ResponseWriter, r *http.Request, f sdkservices.ListEventsFilter) (list, error) {
	if f.Limit <= 0 {
		f.Limit = getQueryNum(r, "events_limit", 50)
	}

	if f.MinSequenceNumber == 0 {
		f.MinSequenceNumber = uint64(getQueryNum(r, "events_min_seq", 0))
	}

	sdkCs, err := s.Events().List(r.Context(), f)
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

func (s *svc) events(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listEvents(w, r, sdkservices.ListEventsFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "events", ts)
}

func (s *svc) event(w http.ResponseWriter, r *http.Request) {
	eid, err := sdktypes.ParseEventID(r.PathValue("eid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkE, err := s.Events().Get(r.Context(), eid)
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

	actives, err := s.Deployments().List(r.Context(), sdkservices.ListDeploymentsFilter{
		State: sdktypes.DeploymentStateActive,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	testings, err := s.Deployments().List(r.Context(), sdkservices.ListDeploymentsFilter{
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

	projects, err := s.Projects().List(r.Context(), sdktypes.InvalidOrgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// name -> deployment id
	deployments := kittehs.ListToMap(append(actives, testings...), func(d sdktypes.Deployment) (string, string) {
		did := d.ID().String()

		name := did

		_, p := kittehs.FindFirst(projects, func(p sdktypes.Project) bool { return p.ID() == d.ProjectID() })
		if p.IsValid() {
			name = p.Name().String() + "/" + name
		}

		return fmt.Sprintf("%s [%v]", name, d.State()), did
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

func (s *svc) redispatchEvent(w http.ResponseWriter, r *http.Request) {
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
		DeploymentID: did,
	}

	eid1, err := s.Dispatcher().Redispatch(r.Context(), eid, &opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, rootPath+"events/"+eid1.String(), http.StatusSeeOther)
}
