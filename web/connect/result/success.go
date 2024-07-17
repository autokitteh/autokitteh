package result

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed success.html
var successPage string

var successTmpl = template.Must(template.New("success").Parse(successPage))

func (h handler) Success(w http.ResponseWriter, r *http.Request) {
	h.ServeHTTP(w, r, successTmpl)
}
