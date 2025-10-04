package connect

import (
	_ "embed"
	"html/template"
	"net/http"

	"go.jetify.com/typeid"

	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
)

//go:embed index.html
var index string

var tmpl = template.Must(template.New("index").Parse(index))

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Secrecy isn't needed here because every request from GitHub to
	// this webhook will be signed and verified. The only requirement
	// is uniqueness, which TypeID guarantees (UUIDv7 contains both
	// a millisecond-precision timestamp and a random value).
	random := typeid.Must(typeid.WithPrefix(""))
	data := map[string]string{
		"address": fixtures.ServiceAddress(),
		"path":    random.String(),
	}
	if err := tmpl.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
