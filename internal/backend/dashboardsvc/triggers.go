package dashboardsvc

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (s *svc) initTriggers() {
	s.HandleFunc(rootPath+"triggers", s.triggers)
	s.HandleFunc(rootPath+"triggers/{tid}", s.trigger)
}

type trigger struct{ sdktypes.Trigger }

func (p trigger) FieldsOrder() []string {
	return []string{"trigger_id", "name", "connection_id", "project_id"}
}

func (p trigger) HideFields() []string        { return nil }
func (p trigger) ExtraFields() map[string]any { return nil }

func toTrigger(sdkP sdktypes.Trigger) trigger { return trigger{sdkP} }

func (s *svc) listTriggers(w http.ResponseWriter, r *http.Request, f sdkservices.ListTriggersFilter) (list, error) {
	sdkCs, err := s.Triggers().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	var drops []string

	if f.ProjectID.IsValid() {
		drops = append(drops, "project_id")
	}

	return genListData(f, kittehs.Transform(sdkCs, toTrigger), drops...), nil
}

func (s *svc) triggers(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listTriggers(w, r, sdkservices.ListTriggersFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "triggers", ts)
}

func (s *svc) trigger(w http.ResponseWriter, r *http.Request) {
	tid, err := sdktypes.StrictParseTriggerID(r.PathValue("tid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkP, err := s.Triggers().Get(r.Context(), tid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := toTrigger(sdkP)

	renderObject(w, r, "trigger", p.ToProto())
}
