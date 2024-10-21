package httpsvc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"

	"connectrpc.com/grpcreflect"
	"github.com/rs/cors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"go.autokitteh.dev/autokitteh/internal/backend/telemetry"
)

type Svc interface {
	MainURL() string
	MainMux() *http.ServeMux
	MainAddr() string // available only after start.

	AuxMux() *http.ServeMux
	AuxAddr() string // available only after start.
}

type svc struct {
	mainURL           string
	mainMux, auxMux   *http.ServeMux
	mainAddr, auxAddr string
}

func (s *svc) MainURL() string         { return s.mainURL }
func (s *svc) MainMux() *http.ServeMux { return s.mainMux }
func (s *svc) AuxMux() *http.ServeMux  { return s.auxMux }
func (s *svc) MainAddr() string        { return s.mainAddr }
func (s *svc) AuxAddr() string         { return s.auxAddr }

func New(lc fx.Lifecycle, l *zap.Logger, cfg *Config, reflectors []string, extractors []RequestLogExtractor, telemetry *telemetry.Telemetry) (Svc, error) {
	cors := cors.New(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
	})

	interceptedMainMux := http.NewServeMux()

	mainInterceptor, err := intercept(l.Named("main_interceptor"), &cfg.Logger, extractors, interceptedMainMux, telemetry)
	if err != nil {
		return nil, fmt.Errorf("interceptor: %w", err)
	}

	if cfg.EnableGRPCReflection {
		reflector := grpcreflect.NewStaticReflector(reflectors...)
		interceptedMainMux.Handle(grpcreflect.NewHandlerV1(reflector))
		interceptedMainMux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}

	mainServer := http.Server{
		Addr:    cfg.Addr,
		Handler: cors.Handler(mainInterceptor),
	}

	// TODO(ENG-43): Do we need H2C?
	if cfg.H2C.Enable {
		l.Debug("using h2c")
		mainServer.Handler = h2c.NewHandler(mainServer.Handler, &http2.Server{})
	}

	svc := &svc{
		mainMux: interceptedMainMux,
		mainURL: cfg.ServiceURL,
	}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			l := l.Named("main_server")

			ln, err := net.Listen("tcp", mainServer.Addr)
			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			svc.mainAddr = ln.Addr().String()
			if host, port, err := net.SplitHostPort(svc.mainAddr); err == nil {
				ip := net.ParseIP(host)

				var addr string
				if ip.IsUnspecified() {
					addr = "localhost"
				} else {
					addr = ip.To4().String()
				}

				svc.mainAddr = fmt.Sprintf("%s:%s", addr, port)
			}

			l.Info("listening", zap.String("addr", svc.mainAddr))

			if cfg.AddrFilename != "" {
				if err := os.WriteFile(cfg.AddrFilename, []byte(svc.mainAddr), 0o600); err != nil {
					l.Panic("write to addr file failed", zap.Error(err), zap.String("filename", cfg.AddrFilename))
				}
			}

			go func() {
				if err := mainServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
					l.Panic("server error", zap.Error(err))
				}

				l.Debug("server closed")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			l.Named("main_server").Info("shutting down")
			return mainServer.Shutdown(ctx)
		},
	})

	interceptedAuxMux := http.NewServeMux()

	auxInterceptor, err := intercept(l.Named("aux_interceptor"), &cfg.Logger, extractors, interceptedAuxMux, telemetry)
	if err != nil {
		return nil, fmt.Errorf("interceptor: %w", err)
	}

	auxServer := http.Server{
		Addr:    cfg.AuxAddr,
		Handler: cors.Handler(auxInterceptor),
	}

	if cfg.H2C.Enable {
		auxServer.Handler = h2c.NewHandler(auxServer.Handler, &http2.Server{})
	}

	svc.auxMux = interceptedAuxMux

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			l := l.Named("aux_server")

			ln, err := net.Listen("tcp", auxServer.Addr)
			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			svc.auxAddr = ln.Addr().String()
			if host, port, err := net.SplitHostPort(svc.auxAddr); err == nil {
				ip := net.ParseIP(host)

				var addr string
				if ip.IsUnspecified() {
					addr = "localhost"
				} else {
					addr = ip.To4().String()
				}

				svc.auxAddr = fmt.Sprintf("%s:%s", addr, port)
			}

			l.Info("listening", zap.String("addr", svc.auxAddr))

			go func() {
				if err := auxServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
					l.Panic("server error", zap.Error(err))
				}

				l.Debug("server closed")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			l.Named("aux_server").Info("shutting down")
			return auxServer.Shutdown(ctx)
		},
	})

	return svc, nil
}
