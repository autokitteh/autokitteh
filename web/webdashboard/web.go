package webdashboard

import (
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/Masterminds/sprig/v3"

	"go.autokitteh.dev/autokitteh/internal/backend/auth/authcontext"
	"go.autokitteh.dev/autokitteh/internal/backend/fixtures"
	"go.autokitteh.dev/autokitteh/internal/backend/httpsvc"
	"go.autokitteh.dev/autokitteh/internal/version"
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

				return u.Title()
			},
			"ProcessID": func() any { return fixtures.ProcessID() },
			"Version":   func() any { return version.Version },
			"Uptime":    func() any { return fixtures.Uptime().Truncate(time.Second) },
			"Duration":  func() time.Duration { return time.Since(httpsvc.GetStartTime(r.Context())).Truncate(time.Microsecond) },
		}).
		ParseFS(tmplFS, "*.html"))
}
