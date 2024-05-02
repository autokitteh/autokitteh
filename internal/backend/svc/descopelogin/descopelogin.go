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

func IndexPage(w http.ResponseWriter, r *http.Request, projectID string) {
	if err := indexTemplate.Execute(w, projectID); err != nil {
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
	}
}

func LoginPage(w http.ResponseWriter, r *http.Request, projectID string) {
	if err := loginTemplate.Execute(w, projectID); err != nil {
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
	}
}
