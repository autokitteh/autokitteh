package dashboardsvc

import (
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s Svc) initDeployments() {
	s.Muxes.AuthHandleFunc("/deployments", s.deployments)
	s.Muxes.AuthHandleFunc("/deployments/{did}", s.deployment)
}

type deployment struct{ sdktypes.Deployment }

func (p deployment) FieldsOrder() []string {
	return []string{"deployment_id", "name"}
}

func (p deployment) HideFields() []string { return nil }

func toDeployment(sdkD sdktypes.Deployment) deployment { return deployment{sdkD} }

func (s Svc) listDeployments(w http.ResponseWriter, r *http.Request, f sdkservices.ListDeploymentsFilter) (list, error) {
	sdkCs, err := s.Svcs.Deployments().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(kittehs.Transform(sdkCs, toDeployment)), nil
}

func (s Svc) deployments(w http.ResponseWriter, r *http.Request) {
	l, err := s.listDeployments(w, r, sdkservices.ListDeploymentsFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "deployments", l)
}

func (s Svc) deployment(w http.ResponseWriter, r *http.Request) {
	did, err := sdktypes.ParseDeploymentID(r.PathValue("did"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkD, err := s.Svcs.Deployments().Get(r.Context(), did)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sess, err := s.listSessions(w, r, sdkservices.ListSessionsFilter{
		DeploymentID: did,
	})
	if err != nil {
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "deployment.html", struct {
		Message  string
		Title    string
		ID       string
		JSON     template.HTML
		Sessions list
	}{
		Title:    "Deployment: " + sdkD.ID().String(),
		ID:       sdkD.ID().String(),
		JSON:     marshalObject(sdkD.ToProto()),
		Sessions: sess,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
