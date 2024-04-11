package svc

import (
	_ "embed"
	"net/http"
	"text/template"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/kittehs"
)

//go:embed index.html
var indexContent string

var indexTemplate = kittehs.Must1(template.New("ee_index.html").Parse(indexContent))

func indexOption() fx.Option {
	return fx.Invoke(func(z *zap.Logger, mux *http.ServeMux) {
		mux.Handle("/{$}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			kittehs.Must0(indexTemplate.Execute(w, nil))
		}))
	})
}
