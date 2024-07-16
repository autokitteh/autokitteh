package dashboardsvc

import (
	"encoding/json"
	"errors"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initBuilds() {
	s.Muxes.Auth.HandleFunc("/builds", s.builds)
	s.Muxes.Auth.HandleFunc("/builds/{bid}", s.build)
}

type build struct{ sdktypes.Build }

func (p build) FieldsOrder() []string       { return nil }
func (p build) HideFields() []string        { return nil }
func (p build) ExtraFields() map[string]any { return nil }

func toBuild(sdkP sdktypes.Build) build { return build{sdkP} }

func (s Svc) listBuilds(w http.ResponseWriter, r *http.Request) (list, error) {
	f := sdkservices.ListBuildsFilter{}

	sdkCs, err := s.Svcs.Builds().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(f, kittehs.Transform(sdkCs, toBuild)), nil
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

	sdkB, err := s.Svcs.Builds().Get(r.Context(), bid)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdkerrors.ErrNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}

	bf, err := s.Svcs.Builds().Describe(r.Context(), bid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "build.html", struct {
		Title string
		ID    string
		Build template.HTML
		File  template.HTML
	}{
		Title: "Build: " + sdkB.ID().String(),
		ID:    bid.String(),
		Build: marshalObject(sdkB.ToProto()),
		File:  template.HTML(kittehs.Must1(json.MarshalIndent(bf, "", "  "))),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
