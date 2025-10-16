package dashboardsvc

import (
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdktypes"
	"go.autokitteh.dev/autokitteh/web/webdashboard"
)

type org struct{ sdktypes.Org }

func (p org) FieldsOrder() []string       { return []string{"org_id", "name", "display_name"} }
func (p org) HideFields() []string        { return nil }
func (p org) ExtraFields() map[string]any { return nil }

func toOrg(sdkO sdktypes.Org) org { return org{sdkO} }

func (s *svc) initOrgs() {
	s.HandleFunc(rootPath+"orgs", s.orgs)
	s.HandleFunc(rootPath+"orgs/{oid}", s.org)
	s.HandleFunc(rootPath+"orgs/{oid}/projects", s.projects)
}

func (s *svc) orgs(w http.ResponseWriter, r *http.Request) {
	_, os, err := s.Orgs().GetOrgsForUser(r.Context(), sdktypes.InvalidUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	orgs := genListData(nil, kittehs.Transform(os, toOrg))

	renderList(w, r, "orgs", orgs)
}

func (s *svc) org(w http.ResponseWriter, r *http.Request) {
	oid, err := sdktypes.StrictParseOrgID(r.PathValue("oid"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	o, err := s.Orgs().GetByID(r.Context(), oid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ms, _, err := s.Orgs().ListMembers(r.Context(), oid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	members := genListData(nil, kittehs.Transform(ms, toMembership))

	ps, err := s.listProjects(w, r, oid)
	if err != nil {
		return
	}

	if err := webdashboard.Tmpl(r).ExecuteTemplate(w, "org.html", struct {
		Title    string
		JSON     template.HTML
		Projects list
		Members  list
	}{
		Title:    "Org: " + o.ID().String(),
		JSON:     marshalObject(o.ToProto()),
		Projects: ps,
		Members:  members,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
