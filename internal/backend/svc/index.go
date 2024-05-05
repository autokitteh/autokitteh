package svc

import (
	_ "embed"
	"net/http"
	"text/template"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/muxes"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed index.html
var indexContent string

var indexTemplate = kittehs.Must1(template.New("index.html").Parse(indexContent))

func indexOption() fx.Option {
	return fx.Invoke(func(z *zap.Logger, muxes *muxes.Muxes) {
		muxes.Handle("/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		}))
	})
}
