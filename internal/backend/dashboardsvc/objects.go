package dashboardsvc

import (
	"net/http"
	"strings"
)

var routes = map[string]string{
	"prj": "projects",
	"con": "connections",
	"int": "integrations",
	"trg": "triggers",
	"env": "envs",
	"ses": "sessions",
	"dep": "deployments",
	"bld": "builds",
	"evt": "events",
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
