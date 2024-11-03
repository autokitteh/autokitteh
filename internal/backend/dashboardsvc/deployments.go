package dashboardsvc

import (
	"context"
	"errors"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
	"go.autokitteh.dev/autokitteh/sdk/sdkservices"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

func (s *svc) initDeployments() {
	deployments := s.Deployments()

	s.HandleFunc(rootPath+"deployments", s.deployments)
	s.HandleFunc(rootPath+"deployments/{did}", s.deployment)
	s.HandleFunc(rootPath+"deployments/{did}/activate", s.deploymentAction(deployments.Activate))
	s.HandleFunc(rootPath+"deployments/{did}/deactivate", s.deploymentAction(deployments.Deactivate))
	s.HandleFunc(rootPath+"deployments/{did}/test", s.deploymentAction(deployments.Test))
}

type deployment struct{ sdktypes.Deployment }

func (p deployment) FieldsOrder() []string {
	return []string{"deployment_id", "name"}
}

func (p deployment) HideFields() []string { return nil }

func (p deployment) ExtraFields() map[string]any { return nil }

func toDeployment(sdkD sdktypes.Deployment) deployment { return deployment{sdkD} }

func (s *svc) listDeployments(w http.ResponseWriter, r *http.Request, f sdkservices.ListDeploymentsFilter) (list, error) {
	f.Limit = uint32(getQueryNum(r, "deployments_limit", 50))

	sdkCs, err := s.Deployments().List(r.Context(), f)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return list{}, err
	}

	return genListData(f, kittehs.Transform(sdkCs, toDeployment)), nil
}

func (s *svc) deployments(w http.ResponseWriter, r *http.Request) {
	l, err := s.listDeployments(w, r, sdkservices.ListDeploymentsFilter{})
	if err != nil {
		return
	}

	renderList(w, r, "deployments", l)
}

func (s *svc) deployment(w http.ResponseWriter, r *http.Request) {
	did, err := sdktypes.ParseDeploymentID(r.PathValue("did"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sdkD, err := s.Deployments().Get(r.Context(), did)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, sdkerrors.ErrNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
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
		State    string
	}{
		Title:    "Deployment: " + sdkD.ID().String(),
		ID:       sdkD.ID().String(),
		JSON:     marshalObject(sdkD.ToProto()),
		Sessions: sess,
		State:    sdkD.State().String(),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *svc) deploymentAction(act func(context.Context, sdktypes.DeploymentID) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		did, err := sdktypes.Strict(sdktypes.ParseDeploymentID(r.PathValue("did")))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := act(r.Context(), did); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, rootPath+"deployments/"+did.String(), http.StatusSeeOther)
	}
}
