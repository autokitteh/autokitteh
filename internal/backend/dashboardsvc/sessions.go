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

func (s *svc) initSessions() {
	s.HandleFunc(rootPath+"sessions", s.sessions)
	s.HandleFunc(rootPath+"sessions/{sid}", s.session)
	s.HandleFunc(rootPath+"sessions/{sid}/stop", s.stopSession)
}

type session struct{ sdktypes.Session }

func (p session) FieldsOrder() []string {
	return []string{"created_at", "session_id", "name", "connection_id", "env_id"}
}

func (p session) HideFields() []string        { return nil }
func (p session) ExtraFields() map[string]any { return nil }

func toSession(sdkP sdktypes.Session) session { return session{sdkP} }

func (s *svc) listSessions(w http.ResponseWriter, r *http.Request, f sdkservices.ListSessionsFilter) (list, error) {
	sdkCs, err := s.Sessions().List(r.Context(), f)
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

func (s *svc) sessions(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listSessions(w, r, sdkservices.ListSessionsFilter{
		PaginationRequest: sdktypes.PaginationRequest{
			PageSize:  int32(getQueryNum(r, "sessions_page_size", 50)),
			Skip:      int32(getQueryNum(r, "sessions_skip", 0)),
			PageToken: r.URL.Query().Get("sessions_page_token"),
		},
	})
	if err != nil {
		return
	}

	renderList(w, r, "sessions", ts)
}

func (s *svc) session(w http.ResponseWriter, r *http.Request) {
	sid, err := sdktypes.StrictParseSessionID(r.PathValue("sid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkS, err := s.Sessions().Get(r.Context(), sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log, err := s.Sessions().GetLog(r.Context(), sdkservices.ListSessionLogRecordsFilter{SessionID: sid})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vw := sdktypes.DefaultValueWrapper
	vw.SafeForJSON = true
	vw.IgnoreFunctions = true

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

	for _, r := range log.Log.Records() {
		if s, ok := r.GetPrint(); ok {
			prints += s + "\n"
		}
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "session.html", struct {
		Title       string
		ID          string
		SessionJSON template.HTML
		LogJSON     template.HTML
		LogText     []template.HTML
		InputsJSON  template.HTML
		Prints      string
		State       string
		IsActive    bool
	}{
		Title:       "Session: " + sdkS.ID().String(),
		ID:          sdkS.ID().String(),
		SessionJSON: marshalObject(sdkS.WithInputs(nil).ToProto()),
		LogJSON:     template.HTML(kittehs.Must1(kittehs.MarshalProtoSliceJSON(log.Log.ToProto().Records))),
		LogText: kittehs.Transform(log.Log.Records(), func(r sdktypes.SessionLogRecord) template.HTML {
			return template.HTML(r.ToString())
		}),
		InputsJSON: template.HTML(jsonInputs),
		Prints:     prints,
		State:      sdkS.State().String(),
		IsActive:   !sdkS.State().IsFinal(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *svc) stopSession(w http.ResponseWriter, r *http.Request) {
	sid, err := sdktypes.StrictParseSessionID(r.PathValue("sid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	reason := r.URL.Query().Get("reason")
	force := getQueryBool(r, "force", false)

	if err := s.Sessions().Stop(r.Context(), sid, reason, force); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf(rootPath+"/sessions/%v", sid), http.StatusSeeOther)
}
