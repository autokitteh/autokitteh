package dashboardsvc

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (s Svc) initTriggers() {
	s.Muxes.Auth.HandleFunc("/triggers", s.triggers)
	s.Muxes.Auth.HandleFunc("/triggers/{tid}", s.trigger)
}

type trigger struct{ sdktypes.Trigger }

func (p trigger) FieldsOrder() []string {
	return []string{"trigger_id", "name", "connection_id", "env_id"}
}

func (p trigger) HideFields() []string { return nil }

func toTrigger(sdkP sdktypes.Trigger) trigger { return trigger{sdkP} }

func (s Svc) listTriggers(w http.ResponseWriter, r *http.Request, f sdkservices.ListTriggersFilter) (list, error) {
	sdkCs, err := s.Svcs.Triggers().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	var drops []string
	if f.EnvID.IsValid() {
		drops = append(drops, "env_id")
	}

	if f.ProjectID.IsValid() {
		drops = append(drops, "project_id")
	}

	return genListData(kittehs.Transform(sdkCs, toTrigger), drops...), nil
}

func (s Svc) triggers(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listTriggers(w, r, sdkservices.ListTriggersFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "triggers", ts)
}

func (s Svc) trigger(w http.ResponseWriter, r *http.Request) {
	tid, err := sdktypes.StrictParseTriggerID(r.PathValue("tid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkP, err := s.Svcs.Triggers().Get(r.Context(), tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := toTrigger(sdkP)

	renderObject(w, r, "trigger", p.ToProto())
}
