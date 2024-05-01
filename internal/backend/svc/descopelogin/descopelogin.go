package descopelogin

import (
	_ "embed"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed index.html
var index string
var indexTemplate = kittehs.Must1(template.New("index.html").Parse(index))

//go:embed login.html
var login string
var loginTemplate = kittehs.Must1(template.New("index.html").Parse(login))

func IndexPage(w http.ResponseWriter, projectID string) {
	indexTemplate.Execute(w, projectID)
}

func LoginPage(w http.ResponseWriter, projectID string) {
	loginTemplate.Execute(w, projectID)
}
