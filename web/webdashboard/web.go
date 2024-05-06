package webdashboard

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/Masterminds/sprig/v3"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
)

//go:embed *.html
var tmplFS embed.FS

func Tmpl(r *http.Request) *template.Template {
	return template.Must(template.New("").
		Funcs(sprig.FuncMap()).
		Funcs(map[string]any{
			"User": func() any {
				u := authcontext.GetAuthnUser(r.Context())
				if !u.IsValid() {
					return nil
				}

				return u.UniqueID()
			},
		}).
		ParseFS(tmplFS, "*.html"))
}
