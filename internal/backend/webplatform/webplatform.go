package webplatform

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"go.autokitteh.dev/autokitteh/internal/backend/configset"
	"go.autokitteh.dev/autokitteh/internal/kittehs"
	"go.autokitteh.dev/autokitteh/sdk/sdkerrors"
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

	addr    string
	l       *zap.Logger
	srv     *http.Server
	version string
}

func (w *Svc) Addr() string    { return w.addr }
func (w *Svc) Version() string { return w.version }

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

	webfs, version, err := webplatform.LoadFS(w.l)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrNotFound) {
			w.l.Warn("web platform distribution not found, web platform server disabled. Run `make ak` to build with the latest webplatform.")
			return nil
		}
		return err
	}

	l := w.l.With(zap.String("version", version), zap.Int("port", w.Config.Port))
	w.version = version

	fsrv := http.FileServer(http.FS(webfs))

	srv := &http.Server{
		Addr: kittehs.BindingAddress(strconv.Itoa(w.Config.Port)),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If path actually exists in fs, serve it.
			if f, _ := webfs.Open(strings.TrimPrefix(r.URL.Path, "/")); f != nil {
				f.Close()
				fsrv.ServeHTTP(w, r)
				return
			}

			// Otherwise redirect all other queries to /index.html.
			http.ServeFileFS(w, r, webfs, "/index.html")
		}),
	}

	w.addr = kittehs.DisplayAddress(srv.Addr)

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
