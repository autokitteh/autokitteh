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
	Dev:     &Config{Port: 9990},
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

	fs, err := webplatform.LoadFS()
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", w.Config.Port),
		Handler: http.FileServer(http.FS(fs)),
	}

	w.addr = srv.Addr

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			w.l.Error("web platform server error", zap.Error(err))
			return
		}

		w.l.Info("web platform server stopped")
	}()

	w.l.Info("web platform server started", zap.String("addr", w.addr))

	return nil
}

func (w *Svc) Stop(context.Context) error {
	if w.srv == nil {
		return nil
	}

	return w.srv.Close()
}
