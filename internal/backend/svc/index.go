package svc

import (
	_ "embed"
	"net/http"
	"text/template"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/auth"
	"go.autokitteh.dev/autokitteh/internal/backend/svc/descopelogin"
	"go.autokitteh.dev/autokitteh/internal/backend/svc/errorpage"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed index.html
var indexContent string

var indexTemplate = kittehs.Must1(template.New("index.html").Parse(indexContent))

func indexOption() fx.Option {
	return fx.Invoke(func(z *zap.Logger, mux *http.ServeMux, authenticator auth.Authenticator) {
		mux.Handle("/error", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			errorpage.ErrorPage(w, "test")
		}))

		mux.Handle("/login", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch authenticator.Provider().Name {

			case auth.AuthProviderDescope:
				projectID := authenticator.Provider().Config["ProjectID"]
				descopelogin.LoginPage(w, r, projectID)
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
				descopelogin.IndexPage(w, r, projectID)
				return

			default:
				z.Error("login unknown provider", zap.String("provider", authenticator.Provider().Name))
				http.Redirect(w, r, "/error", 302)
			}
		}))
	})
}
