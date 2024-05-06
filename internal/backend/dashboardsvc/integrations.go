package dashboardsvc

import (
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
)

func (s Svc) initIntegrations() {
	s.Muxes.AuthHandleFunc("/integrations", s.integrations)
	s.Muxes.AuthHandleFunc("/integrations/{iid}", s.integration)
}

type integration struct{ sdktypes.Integration }

func (p integration) FieldsOrder() []string { return []string{"unique_name", "integration_id"} }
func (p integration) HideFields() []string {
	return []string{}
}

func toIntegration(sdkI sdktypes.Integration) integration { return integration{sdkI} }

func (s Svc) listIntegrations(w http.ResponseWriter, r *http.Request) (list, error) {
	sdkIs, err := s.Svcs.Integrations().List(r.Context(), "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(kittehs.Transform(sdkIs, toIntegration)), nil
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

	sdkI, err := s.Svcs.Integrations().GetByID(r.Context(), iid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderObject(w, r, "integration", sdkI.Get().ToProto())
}
