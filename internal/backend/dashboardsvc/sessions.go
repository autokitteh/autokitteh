package dashboardsvc

import (
	"encoding/json"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initSessions() {
	s.Muxes.Auth.HandleFunc("/sessions", s.sessions)
	s.Muxes.Auth.HandleFunc("/sessions/{sid}", s.session)
}

type session struct{ sdktypes.Session }

func (p session) FieldsOrder() []string {
	return []string{"created_at", "session_id", "name", "connection_id", "env_id"}
}

func (p session) HideFields() []string        { return nil }
func (p session) ExtraFields() map[string]any { return nil }

func toSession(sdkP sdktypes.Session) session { return session{sdkP} }

func (s Svc) listSessions(w http.ResponseWriter, r *http.Request, f sdkservices.ListSessionsFilter) (list, error) {
	sdkCs, err := s.Svcs.Sessions().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	var drops []string
	if f.EnvID.IsValid() {
		drops = append(drops, "env_id")
	}

	if f.DeploymentID.IsValid() {
		drops = append(drops, "deployment_id")
	}

	if f.EventID.IsValid() {
		drops = append(drops, "event_id")
	}

	if f.BuildID.IsValid() {
		drops = append(drops, "build_id")
	}

	return genListData(f, kittehs.Transform(sdkCs.Sessions, toSession), drops...), nil
}

func (s Svc) sessions(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listSessions(w, r, sdkservices.ListSessionsFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "sessions", ts)
}

func (s Svc) session(w http.ResponseWriter, r *http.Request) {
	sid, err := sdktypes.StrictParseSessionID(r.PathValue("sid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkS, err := s.Svcs.Sessions().Get(r.Context(), sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log, err := s.Svcs.Sessions().GetLog(r.Context(), sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vw := sdktypes.DefaultValueWrapper
	vw.SafeForJSON = true

	inputs, err := kittehs.TransformMapValuesError(sdkS.Inputs(), vw.Unwrap)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonInputs, err := json.MarshalIndent(inputs, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var prints string

	for _, r := range log.Records() {
		if s, ok := r.GetPrint(); ok {
			prints += s + "\n"
		}
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "session.html", struct {
		Title       string
		ID          string
		SessionJSON template.HTML
		LogJSON     template.HTML
		InputsJSON  template.HTML
		Prints      string
	}{
		Title:       "Session: " + sdkS.ID().String(),
		ID:          sdkS.ID().String(),
		SessionJSON: marshalObject(sdkS.WithInputs(nil).ToProto()),
		LogJSON:     template.HTML(kittehs.Must1(kittehs.MarshalProtoSliceJSON(log.ToProto().Records))),
		InputsJSON:  template.HTML(jsonInputs),
		Prints:      prints,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
