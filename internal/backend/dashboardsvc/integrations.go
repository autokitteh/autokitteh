package dashboardsvc

import (
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initIntegrations() {
	s.Muxes.Auth.HandleFunc("/integrations", s.integrations)
	s.Muxes.Auth.HandleFunc("/integrations/{iid}", s.integration)
}

type integration struct{ sdktypes.Integration }

func (p integration) FieldsOrder() []string       { return []string{"unique_name", "integration_id"} }
func (p integration) HideFields() []string        { return nil }
func (p integration) ExtraFields() map[string]any { return nil }

func toIntegration(sdkI sdktypes.Integration) integration { return integration{sdkI} }

func (s Svc) listIntegrations(w http.ResponseWriter, r *http.Request) (list, error) {
	sdkIs, err := s.Svcs.Integrations().List(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(nil, kittehs.Transform(sdkIs, toIntegration)), nil
}

func (s Svc) integrations(w http.ResponseWriter, r *http.Request) {
	is, err := s.listIntegrations(w, r)
	if err != nil {
		return
	}

	renderList(w, r, "integrations", is)
}

func (s Svc) integration(w http.ResponseWriter, r *http.Request) {
	iid, err := sdktypes.StrictParseIntegrationID(r.PathValue("iid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	intg, err := s.Svcs.Integrations().GetByID(r.Context(), iid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "integration.html", struct {
		Title           string
		ID              string
		IntegrationJSON template.HTML
		VarsJSON        template.HTML
		FuncsJSON       template.HTML
	}{
		Title:           "Integration: " + intg.ID().String(),
		ID:              intg.ID().String(),
		IntegrationJSON: marshalObject(intg.WithModule(sdktypes.InvalidModule).ToProto()),
		VarsJSON:        template.HTML(kittehs.Must1(kittehs.MarshalProtoMapJSON(intg.ToProto().Module.Variables))),
		FuncsJSON:       template.HTML(kittehs.Must1(kittehs.MarshalProtoMapJSON(intg.ToProto().Module.Functions))),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
