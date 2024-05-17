package dashboardsvc

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
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

	return genListData(kittehs.Transform(sdkCs.Sessions, toSession), drops...), nil
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

	sdkP, err := s.Svcs.Sessions().Get(r.Context(), sid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := toSession(sdkP)

	renderBigObject(w, r, "session", p.ToProto())
}
