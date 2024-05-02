package svc

import (
	_ "embed"
	"net/http"
	"text/template"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth"
	"go.autokitteh.dev/autokitteh/web/static"

	"go.autokitteh.dev/autokitteh/internal/backend/svc/errorpage"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed index.html
var indexContent string

var indexTemplate = kittehs.Must1(template.New("index.html").Parse(indexContent))

var descopeIndexTemplate = kittehs.Must1(template.New("index.html").Parse(static.DescopeIndex))
var descopeLoginTemplate = kittehs.Must1(template.New("login.html").Parse(static.DescopeLogin))

func descopeIndexPage(w http.ResponseWriter, r *http.Request, projectID string) {
	if err := descopeIndexTemplate.Execute(w, projectID); err != nil {
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
	}
}

func descopeLoginPage(w http.ResponseWriter, r *http.Request, projectID string) {
	if err := descopeLoginTemplate.Execute(w, projectID); err != nil {
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
	}
}

func indexOption() fx.Option {
	return fx.Invoke(func(z *zap.Logger, mux *http.ServeMux, authenticator auth.Authenticator) {
		mux.Handle("/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errorpage.ErrorPage(w, "test")
		}))

		mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch authenticator.Provider().Name {

			case auth.AuthProviderDescope:
				projectID := authenticator.Provider().Config["ProjectID"]
				descopeLoginPage(w, r, projectID)
				return

			default:
				z.Error("login unknown provider", zap.String("provider", authenticator.Provider().Name))
				http.Redirect(w, r, "/error", 302)
			}
		}))

		mux.Handle("/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch authenticator.Provider().Name {

			case auth.AuthProviderNone:
				kittehs.Must0(indexTemplate.Execute(w, nil))
				return

			case auth.AuthProviderDescope:
				projectID := authenticator.Provider().Config["ProjectID"]
				descopeIndexPage(w, r, projectID)
				return

			default:
				z.Error("login unknown provider", zap.String("provider", authenticator.Provider().Name))
				http.Redirect(w, r, "/error", 302)
			}
		}))
	})
}
