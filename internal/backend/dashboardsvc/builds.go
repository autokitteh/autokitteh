package dashboardsvc

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (s Svc) initBuilds() {
	s.Muxes.AuthHandleFunc("/builds", s.builds)
	s.Muxes.AuthHandleFunc("/builds/{bid}", s.build)
}

type build struct{ sdktypes.Build }

func (p build) FieldsOrder() []string { return nil }
func (p build) HideFields() []string  { return nil }

func toBuild(sdkP sdktypes.Build) build { return build{sdkP} }

func (s Svc) listBuilds(w http.ResponseWriter, r *http.Request) (list, error) {
	sdkCs, err := s.Svcs.Builds().List(r.Context(), sdkservices.ListBuildsFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(kittehs.Transform(sdkCs, toBuild)), nil
}

func (s Svc) builds(w http.ResponseWriter, r *http.Request) {
	ts, err := s.listBuilds(w, r)
	if err != nil {
		return
	}

	renderList(w, r, "builds", ts)
}

func (s Svc) build(w http.ResponseWriter, r *http.Request) {
	bid, err := sdktypes.StrictParseBuildID(r.PathValue("bid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkP, err := s.Svcs.Builds().Get(r.Context(), bid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	p := toBuild(sdkP)

	renderObject(w, r, "build", p.ToProto())
}
