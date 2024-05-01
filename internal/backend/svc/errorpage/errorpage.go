package errorpage

import (
	_ "embed"
	"html/template"
	"net/http"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed error.html
var errorPage string

var errorTemplate = kittehs.Must1(template.New("index.html").Parse(errorPage))

func ErrorPage(w http.ResponseWriter, message string) {
	errorTemplate.Execute(w, message)
}
