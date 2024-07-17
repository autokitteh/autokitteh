package result

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed error.html
var errorPage string

var errorTmpl = template.Must(template.New("success").Parse(errorPage))

func (h handler) Error(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r, errorTmpl)
}
