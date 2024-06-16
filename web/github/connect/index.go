package connect

import (
	_ "embed"
	"html/template"
	"net/http"
	"os"

	"go.jetify.com/typeid"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed index.html
var index string

var tmpl = template.Must(template.New("index").Parse(index))

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	random := kittehs.Must1(typeid.WithPrefix(""))
	data := map[string]string{
		"address": os.Getenv("WEBHOOK_ADDRESS"),
		"path":    random.String(),
	}
	if err := tmpl.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
