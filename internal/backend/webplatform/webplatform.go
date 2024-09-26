package webplatform

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/web/webplatform"
)

const defaultPort = 9982

type Config struct {
	Port int `koanf:"port"` // 0 - disabled
}

var Configs = configset.Set[Config]{
	Default: &Config{},
	Dev:     &Config{Port: defaultPort},
	Test:    &Config{},
}

type Svc struct {
	Config *Config

	addr string
	l    *zap.Logger
	srv  *http.Server
}

func (w *Svc) Addr() string { return w.addr }

func New(cfg *Config, l *zap.Logger) *Svc {
	return &Svc{
		Config: cfg,
		l:      l,
	}
}

func (w *Svc) Start(context.Context) error {
	if w.Config.Port == 0 {
		return nil
	}

	fs, version, err := webplatform.LoadFS()
	if err != nil {
		return err
	}

	l := w.l.With(zap.String("version", version), zap.Int("port", w.Config.Port))

	fsrv := http.FileServer(http.FS(fs))

	srv := &http.Server{
		Addr: fmt.Sprintf("localhost:%d", w.Config.Port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If path actually exists in fs, serve it.
			if f, _ := fs.Open(strings.TrimPrefix(r.URL.Path, "/")); f != nil {
				f.Close()
				fsrv.ServeHTTP(w, r)
				return
			}

			// Otherwise redirect all other queries to /index.html.
			http.ServeFileFS(w, r, fs, "/index.html")
		}),
	}

	w.addr = srv.Addr

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			l.Error("web platform server error", zap.Error(err))
			return
		}

		l.Info("web platform server stopped")
	}()

	l.Info("web platform server started")

	return nil
}

func (w *Svc) Stop(context.Context) error {
	if w.srv == nil {
		return nil
	}

	return w.srv.Close()
}
