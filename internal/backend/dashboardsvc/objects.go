package dashboardsvc

import (
	"net/http"
	"strings"
)

var routes = map[string]string{
	"bld": "builds",
	"con": "connections",
	"dep": "deployments",
	"env": "envs",
	"evt": "events",
	"int": "integrations",
	"org": "orgs",
	"prj": "projects",
	"ses": "sessions",
	"trg": "triggers",
	"usr": "users",
}

func (s *svc) initObjects() {
	s.HandleFunc(rootPath+"objects/{id}", s.objects)
}

func (s *svc) objects(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	prefix, _, ok := strings.Cut(id, "_")
	if !ok {
		http.Error(w, "invalid object id", http.StatusBadRequest)
		return
	}

	if dst, ok := routes[prefix]; ok {
		http.Redirect(w, r, rootPath+dst+"/"+id, http.StatusFound)
		return
	}

	http.NotFound(w, r)
}
