package webplatform

import (
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/web/webplatform"
)

const defaultPort = 9990

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

	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", w.Config.Port),
		Handler: http.FileServer(http.FS(fs)),
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
