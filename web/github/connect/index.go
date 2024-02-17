package connect

import (
	_ "embed"
	"html/template"
	"net/http"
	"os"

	"github.com/lithammer/shortuuid/v4"
)

//go:embed index.html
var index string

var tmpl = template.Must(template.New("index").Parse(index))

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"address": os.Getenv("WEBHOOK_ADDRESS"),
		"path":    shortuuid.New(),
	}
	if err := tmpl.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
