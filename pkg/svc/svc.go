package svc

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" // pprof
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/rs/cors"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/autokitteh/L"
	"github.com/autokitteh/L/Z"
	"github.com/autokitteh/flexcall"
)

var DefaultServiceName = filepath.Base(os.Args[0])

type svc struct{ opts opts }

func Run(opts ...OptFunc) {
	if err := <-MustStart(opts...); err != nil {
		panic(err)
	}
}

func MustStart(opts ...OptFunc) <-chan error {
	ch, err := Start(opts...)
	if err != nil {
		panic(err)
	}

	return ch
}

func call(f interface{}, providers *Providers, l L.L) error {
	outs, err := flexcall.CallOptionalAndExtractError(f, append(providers.Vs, l)...)
	if err != nil {
		return err
	}

	providers.Add(outs...)

	return nil
}

type stringsListFlag []string

func (i *stringsListFlag) String() string {
	return strings.Join(*i, ",")
}

func (i *stringsListFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func parseFlags() *Flags {
	var enables, disables, onlys, excepts stringsListFlag

	flag.Var(&enables, "enable", "modules to enable")
	flag.Var(&disables, "disable", "modules to disable")
	flag.Var(&onlys, "only", "enable only these modules")
	flag.Var(&excepts, "except", "disable only these modules")

	cfgPathFlag := flag.String("config", "", "use config file")
	setupFlag := flag.Bool("setup", false, "run setup pahse")
	helpConfigFlag := flag.Bool("help-config", false, "describe accepted environment variables and exit")
	printConfigFlag := flag.Bool("print-config", false, "print config")
	exitBeforeStartFlag := flag.Bool("exit-before-start", false, "exit before start")

	flag.Parse()

	return &Flags{
		ConfigPath:      *cfgPathFlag,
		Enables:         enables,
		Disables:        disables,
		Excepts:         excepts,
		Onlys:           onlys,
		Setup:           *setupFlag,
		HelpConfig:      *helpConfigFlag,
		PrintConfig:     *printConfigFlag,
		ExitBeforeStart: *exitBeforeStartFlag,
	}
}

// Start the service. Returns a channel for asynchronous error handling, for example
// in case of a failure of the GRPC server itself.
//
// The startup is done in the following order:
// 1. Configuration loading (both user specified and service);
// 2. Log initializations.
// 3. Call user Init functions.
// 4. If -setup is specified, call user Setup functions.
// 5. If -exit-before-start is not specified, call user Start functions.
func Start(opts ...OptFunc) (<-chan error, error) {
	var svc svc

	for _, opt := range opts {
		opt(&svc.opts)
	}

	flags := svc.opts.flags
	if flags == nil {
		flags = parseFlags()
	}

	c := 0
	if len(flags.Onlys) != 0 {
		c++
	}
	if len(flags.Excepts) != 0 {
		c++
	}

	if c > 1 {
		return nil, errors.New("--only and --excepts are mutually exclusive")
	}

	moduleFilter := func(n string) bool {
		for _, e := range flags.Enables {
			if n == e {
				return true
			}
		}

		for _, e := range flags.Disables {
			if n == e {
				return false
			}
		}

		for _, e := range svc.opts.defaultDisables {
			if n == e {
				return false
			}
		}

		if onlys := flags.Onlys; len(onlys) != 0 {
			for _, e := range onlys {
				if n == e {
					return true
				}
			}

			return false
		}

		if excepts := flags.Excepts; len(excepts) != 0 {
			for _, e := range excepts {
				if n == e {
					return false
				}
			}

			return true
		}

		return true
	}

	modulesFilter := func(cbs []callback) (enabled, disabled []callback) {
		enabled = make([]callback, 0, len(cbs))
		disabled = make([]callback, 0, len(cbs))

		for _, cb := range cbs {
			if moduleFilter(cb.n) {
				enabled = append(enabled, cb)
			} else {
				disabled = append(disabled, cb)
			}
		}

		return
	}

	name := DefaultServiceName

	if svc.opts.name != "" {
		name = svc.opts.name
	}

	errCh := make(chan error, 1)

	if flags.HelpConfig {
		printUsage(name, svc.opts.cfgs...)

		errCh <- nil
		return errCh, nil
	}

	providers := &Providers{Vs: svc.opts.providers}
	providers.Add(providers)

	cfg, err := loadSvcCfg(name, flags.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load svc cfg error: %w", err)
	}

	providers.Add(cfg)

	var l L.L

	if f := svc.opts.l; f != nil {
		l = L.N(f())
	} else {
		l, err = Z.NewL(cfg.Log, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("init log error: %w", err)
		}
	}

	l.Debug("log init")

	if v := GetVersion(); v != nil {
		l.Info("initializing", "version", v)
	}

	for _, c := range svc.opts.cfgs {
		if err := loadCfg(l.Named("configs"), name, c, flags.ConfigPath); err != nil {
			return nil, fmt.Errorf("load user cfg error: %w", err)
		}

		providers.Add(c)
	}

	if flags.PrintConfig {
		l.Info("configs", "svc_cfg", cfg, "user_cfgs", svc.opts.cfgs)
	}

	if cfg.PprofPort != 0 {
		l.Debug("starting pprof server", "port", cfg.PprofPort)

		go func() {
			err := http.ListenAndServe(fmt.Sprintf("localhost:%d", cfg.PprofPort), nil)
			l.Errorf("pprof exited", "err", err)
			errCh <- fmt.Errorf("pprof exited: %w", err)
		}()
	}

	ctx := context.Background()

	providers.Add(ctx)

	var grpcOpts GRPCOptions

	if inits, _ := modulesFilter(svc.opts.inits); len(inits) == 0 {
		l.Debug("nothing to initialize")
	} else {
		providers.Add(&grpcOpts)

		l.Debug("initializing", "components", callbacksNames(inits))

		for _, i := range inits {
			l := l.Named(i.n)

			if err := call(i.f, providers, l); err != nil {
				return nil, fmt.Errorf("init error: %w", err)
			}
		}
	}

	if flags.Setup {
		if setups, _ := modulesFilter(svc.opts.setups); len(setups) == 0 {
			l.Info("nothing to setup")
		} else {
			l.Info("setting up", "components", callbacksNames(setups))

			for _, s := range setups {
				l := l.Named(s.n)

				if err := call(s.f, providers, l); err != nil {
					return nil, fmt.Errorf("setup error: %w", err)
				}
			}
		}
	}

	if flags.ExitBeforeStart {
		l.Info("exit before start")
		errCh <- nil

		return errCh, nil
	}

	grpcOpts.Add(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_zap.UnaryServerInterceptor(
					Z.FromL(l.Named("grpc")).Desugar(),
					grpc_zap.WithLevels(func(codes.Code) zapcore.Level { return zapcore.DebugLevel }),
				),
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_zap.StreamServerInterceptor(
					Z.FromL(l.Named("grpc-stream")).Desugar(),
					grpc_zap.WithLevels(func(codes.Code) zapcore.Level { return zapcore.DebugLevel }),
				),
			),
		),
	)

	grpcSrv := grpc.NewServer(append(grpcOpts.opts, grpc.MaxSendMsgSize(1024*1024*50))...)

	providers.Add(grpcSrv)

	httpMux := mux.NewRouter()
	providers.Add(httpMux)

	if starts, _ := modulesFilter(svc.opts.starts); len(starts) == 0 {
		l.Debug("nothing to start")
	} else {
		l.Debug("starting up", "components", callbacksNames(starts))

		for _, s := range starts {
			l := l.Named(s.n)

			if err := call(s.f, providers, l); err != nil {
				return nil, fmt.Errorf("start error: %w", err)
			}
		}
	}

	if svc.opts.grpc && cfg.GRPC.Enabled {
		grpcAddr, err := startGRPC(l.Named("grpc"), grpcSrv, cfg.GRPC, errCh)
		if err != nil {
			return nil, fmt.Errorf("grpc start error: %w", err)
		}

		providers.Add(grpcAddr)
	} else {
		l.Debug("not starting GRPC server")
	}

	if svc.opts.http && cfg.HTTP.Enabled {
		if err := startHTTP(l.Named("http"), httpMux, cfg.HTTP, errCh); err != nil {
			return nil, fmt.Errorf("http start error: %w", err)
		}
	} else {
		l.Debug("not starting HTTP server")
	}

	if readys, _ := modulesFilter(svc.opts.readys); len(readys) == 0 {
		l.Debug("nothing to ready")
	} else {
		l.Debug("readying up", "components", callbacksNames(readys))

		for _, s := range readys {
			l := l.Named(s.n)

			if err := call(s.f, providers, l); err != nil {
				return nil, fmt.Errorf("ready error: %w", err)
			}
		}
	}

	l.Info("ready!")

	return errCh, nil
}

func startGRPC(l L.L, srv *grpc.Server, cfg grpcCfg, errCh chan<- error) (GRPCAddr, error) {
	l.Debug("starting GRPC server", "cfg", cfg)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return nil, fmt.Errorf("grpc listen error: %w", err)
	}

	go func() {
		err := srv.Serve(lis)
		l.Fatal("GRPC serve failed", "err", err)
		errCh <- fmt.Errorf("GRPC serve error: %w", err)
	}()

	if cfg.Port == 0 {
		l.Info("grpc started", "addr", lis.Addr())
	} else {
		l.Debug("grpc started", "addr", lis.Addr())
	}

	return GRPCAddr(lis.Addr()), nil
}

func startHTTP(l L.L, r *mux.Router, cfg httpCfg, errCh chan<- error) error {
	l.Debug("starting HTTP server", "cfg", cfg)

	h := handlers.CombinedLoggingHandler(
		&Z.ApacheLogWriter{
			Z:         Z.FromL(l),
			InfoLevel: cfg.AccessLogInfoLevel,
		},
		r,
	)

	if cfg.CORS {
		h = cors.New(cors.Options{
			AllowedOrigins:   cfg.CORSAllowedOrigins,
			AllowCredentials: cfg.CORSAllowCredentials,
		}).Handler(r)
	}

	go func() {
		err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), h)
		l.Fatal("HTTP serve failed", "err", err)
		errCh <- fmt.Errorf("HTTP serve error: %w", err)
	}()

	return nil
}
