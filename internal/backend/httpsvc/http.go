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
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Svc interface {
	Mux() *http.ServeMux
	Addr() string // available only after start.
}

type svc struct {
	mux  *http.ServeMux
	addr string
}

func (s *svc) Mux() *http.ServeMux { return s.mux }
func (s *svc) Addr() string        { return s.addr }

func New(lc fx.Lifecycle, z *zap.Logger, cfg *Config, reflectors []string, extractors []RequestLogExtractor) (Svc, error) {
	rootMux := http.NewServeMux()

	cors := cors.New(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
	})

	interceptedMux := http.NewServeMux()

	interceptor, err := intercept(z, &cfg.Logger, extractors, interceptedMux)
	if err != nil {
		return nil, fmt.Errorf("interceptor: %w", err)
	}

	rootMux.Handle("/", cors.Handler(interceptor))

	if cfg.EnableGRPCReflection {
		reflector := grpcreflect.NewStaticReflector(reflectors...)
		interceptedMux.Handle(grpcreflect.NewHandlerV1(reflector))
		interceptedMux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))
	}

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: rootMux,
	}

	// TODO(ENG-43): Do we need H2C?
	if cfg.H2C.Enable {
		z.Debug("using h2c")
		server.Handler = h2c.NewHandler(rootMux, &http2.Server{})
	}

	svc := &svc{mux: interceptedMux}

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			var (
				ln  net.Listener
				err error
			)

			if cfg.Ngrok.Enable {
				z.Info("using ngrok to serve apis")

				token := cfg.Ngrok.AuthToken
				if token == "" {
					token = os.Getenv("NGROK_AUTHTOKEN")
				}

				var tun ngrok.Tunnel
				if tun, err = ngrok.Listen(
					context.Background(),
					config.HTTPEndpoint(
						config.WithDomain(cfg.Ngrok.Domain),
					),
					ngrok.WithAuthtoken(token),
				); err == nil {
					ln = tun
				}
			} else {
				ln, err = net.Listen("tcp", server.Addr)
			}

			if err != nil {
				return fmt.Errorf("listen: %w", err)
			}

			svc.addr = ln.Addr().String()
			if host, port, err := net.SplitHostPort(svc.addr); err == nil {
				ip := net.ParseIP(host)

				var addr string
				if ip.IsUnspecified() {
					addr = "localhost"
				} else {
					addr = ip.To4().String()
				}

				svc.addr = fmt.Sprintf("%s:%s", addr, port)
			}

			z.Debug("listening", zap.String("addr", svc.addr))

			if cfg.AddrFilename != "" {
				if err := os.WriteFile(cfg.AddrFilename, []byte(svc.addr), 0o600); err != nil {
					z.Panic("write to addr file failed", zap.Error(err), zap.String("filename", cfg.AddrFilename))
				}
			}

			go func() {
				if err := server.Serve(ln); !errors.Is(http.ErrServerClosed, err) {
					z.Panic("server error", zap.Error(err))
				}

				z.Debug("server closed")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			z.Info("shutting down")
			return server.Shutdown(ctx)
		},
	})

	return svc, nil
}
